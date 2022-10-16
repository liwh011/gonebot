package gonebot

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	goarg "github.com/alexflint/go-arg"
)

type Middleware func(*Context) bool
type HandlerFunc func(*Context)

type Handler struct {
	middlewares []Middleware
	handleFunc  HandlerFunc
	parent      *Handler
	subHandlers map[EventName][]*Handler
	mu          sync.RWMutex
}

// 使用中间件
func (h *Handler) Use(middlewares ...Middleware) *Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.middlewares = append(h.middlewares, middlewares...)
	return h
}

// 指定事件处理函数
func (h *Handler) Handle(f HandlerFunc) {
	h.handleFunc = f
}

// 添加子Handler
func (h *Handler) addSubHandler(subHandler *Handler, eventType ...EventName) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subHandler.parent = h
	for _, event := range eventType {
		h.subHandlers[event] = append(h.subHandlers[event], subHandler)
	}
}

// 移除指定的子Handler
func (h *Handler) removeSubHandler(handler *Handler, eventType ...EventName) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subHandlers == nil {
		return
	}
	for _, event := range eventType {
		for i, subHandler := range h.subHandlers[event] {
			if subHandler == handler {
				h.subHandlers[event] = append(h.subHandlers[event][:i], h.subHandlers[event][i+1:]...)
				return
			}
		}
	}
}

// 新建一个可以被删除的Handler，用于处理指定类型的事件。
//
// 调用remove方法可以删除当前Handler。
func (h *Handler) NewRemovableHandler(eventTypes ...EventName) (handler *Handler, remove func()) {
	handler = &Handler{
		parent:      h,
		subHandlers: make(map[EventName][]*Handler),
	}
	if len(eventTypes) == 0 {
		eventTypes = append(eventTypes, EventName_AllEvent)
	}
	h.addSubHandler(handler, eventTypes...)
	return handler, func() {
		h.removeSubHandler(handler, eventTypes...)
	}
}

// 新建一个Handler，用于处理指定类型的事件
func (h *Handler) NewHandler(eventTypes ...EventName) (handler *Handler) {
	nh, _ := h.NewRemovableHandler(eventTypes...)
	return nh
}

func (h *Handler) getMatchedHandler(eventName EventName) (handlers []*Handler) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 以下构造Handler链，以message.private.friend事件为例，
	// 按message.private.friend、message.private、message、all的顺序将这些Handler放入链中
	parts := strings.Split(string(eventName), ".")
	for i := len(parts); i >= 0; i-- {
		if i == 0 {
			handlers = append(handlers, h.subHandlers[EventName_AllEvent]...)
			break
		}
		shs := h.subHandlers[EventName(strings.Join(parts[:i], "."))]
		handlers = append(handlers, shs...)
	}
	return
}

// 一次事件的处理过程，保存了当前Handler、当前中间件索引、是否中止等信息
type process struct {
	handlerQueue []*Handler // 待的Handler队列

	curHandler  *Handler     // 当前Handler
	isLeaf      bool         // 当前Handler是否是叶子节点
	middlewares []Middleware // 当前Handler的中间件
	mwIdx       int          // 当前正在执行的中间件的索引

	aborted      bool // 是否已经被中断
	next         bool // 是否继续下一个Handler
	done         bool // 当前Handler是否已经执行完毕
	shouldExpand bool // 是否需要展开子Handler

	processedByHandler bool // 是否有Handler处理过当前事件

	ctx       *Context  // 上下文
	eventName EventName // 事件名称
}

// 更新process的当前Handler为队列的下一个，并重置标志
func (proc *process) nextHandler() bool {
	if proc.aborted {
		return false
	}
	if !proc.next {
		return false
	}

	// 当前Handler为非叶子节点，并且中间件未返回false，进行展开
	if !proc.isLeaf && proc.shouldExpand {
		// 按照先序的顺序，子Handler应塞在队头
		proc.handlerQueue = append(proc.curHandler.getMatchedHandler(proc.eventName), proc.handlerQueue...)
	}

	// 队列空，没有了
	if len(proc.handlerQueue) == 0 {
		return false
	}

	proc.curHandler = proc.handlerQueue[0]
	proc.handlerQueue = proc.handlerQueue[1:]
	proc.middlewares = proc.curHandler.middlewares
	proc.isLeaf = len(proc.curHandler.subHandlers) == 0
	proc.mwIdx = 0
	proc.done = false
	proc.shouldExpand = true
	return true
}

// 中止过程
func (proc *process) abort() {
	proc.aborted = true
}

// 运行过程
func (proc *process) run() {
	startHandler := proc.curHandler // 记录开始的Handler
	prevHandler := proc.curHandler  // 上一个Handler，用于比较Handler是否发生变化

	// 当前Handler尚未完成则继续，否则下一个
handlerLoop:
	for !proc.done || proc.nextHandler() {
		if prevHandler != proc.curHandler {
			// 当前Handler发生了变化，需要更新CTX的Handler
			proc.ctx.Handler = proc.curHandler
			prevHandler = proc.curHandler
		}

		// 顺序执行中间件
		for !proc.aborted && proc.mwIdx < len(proc.middlewares) {
			mw := proc.middlewares[proc.mwIdx]
			if !mw(proc.ctx) {
				proc.shouldExpand = false // 中间件返回false，非叶子节点不展开子Handler
				proc.done = true          // 中间件返回false，标志当前Handler执行完毕
				continue handlerLoop      // 执行下一个Handler
			}
			proc.mwIdx++
		}

		if proc.aborted {
			return
		}

		// 如果不是叶子节点，则向后执行它的子Handler
		if !proc.isLeaf {
			proc.done = true // 标志当前Handler执行完毕
			continue
		}

		if proc.curHandler.handleFunc != nil {
			proc.next = false // 叶子节点，默认不向后执行
			// 提前设置done，让下次循环能正确获取下一个Handler。
			// 否则会造成无限递归
			proc.done = true
			proc.curHandler.handleFunc(proc.ctx)
			proc.processedByHandler = true
		}
		proc.done = true
	}

	// 复原CTX的Handler
	proc.ctx.Handler = startHandler
}

// 从当前process创建一个新的process，用于执行当前process的后续
func (proc *process) forkAndNext() bool {
	if proc.aborted {
		return false
	}

	// 为了防止并发调用next，导致两个goroutine同时向后执行，
	// 故规定，调用next之后，原先的process将停止继续处理，转由新的process处理
	proc.aborted = true

	newProc := *proc
	newProc.aborted = false

	if newProc.mwIdx < len(newProc.middlewares) {
		// 调用时，中间件没执行完，则后移一个继续执行
		newProc.mwIdx++
	} else {
		// 调用时，处于处理函数中，则将next置为true
		newProc.next = true
	}

	newProc.ctx.abort = newProc.abort
	newProc.ctx.next = newProc.forkAndNext
	defer func() {
		proc.ctx.abort = proc.abort
		proc.ctx.next = proc.forkAndNext
	}()

	newProc.run()
	proc.processedByHandler = newProc.processedByHandler
	return newProc.processedByHandler
}

func (h *Handler) handleEvent(ctx *Context) {
	proc := process{
		handlerQueue: []*Handler{},
		curHandler:   h,
		isLeaf:       len(h.subHandlers) == 0,
		middlewares:  h.middlewares,
		mwIdx:        0,
		aborted:      false,
		next:         true,
		ctx:          ctx,
		eventName:    ctx.Event.GetEventName(),
		done:         false,
		shouldExpand: true,
	}

	ctx.abort = proc.abort
	ctx.next = proc.forkAndNext

	proc.run()
}

func OnEvent(eventName EventName) Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.GetEventName() == eventName
	}
}

// 与Bot相关
func OnlyToMe() Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.IsToMe()
	}
}

// 限制来自某些群聊，当参数为空时，表示全部群聊都可
func FromGroup(groupIds ...int64) Middleware {
	return func(ctx *Context) bool {
		gid, exist := getEventField(ctx.Event, "GroupId")
		if !exist {
			return false
		}
		if len(groupIds) == 0 {
			return true
		}

		for _, id := range groupIds {
			if id == gid {
				return true
			}
		}
		return false
	}
}

// 限制来自某些人的私聊，当参数为空时，表示只要是私聊都可
func FromPrivate(userIds ...int64) Middleware {
	return func(ctx *Context) bool {
		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			return false
		}
		if len(userIds) == 0 {
			return true
		}

		for _, id := range userIds {
			if id == uid {
				return true
			}
		}
		return false
	}
}

// 消息来源于某些人，必须传入至少一个参数
func FromUser(userIds ...int64) Middleware {
	return func(ctx *Context) bool {
		if len(userIds) == 0 {
			return true
		}

		uid, exist := getEventField(ctx.Event, "UserId")
		if !exist {
			return false
		}
		for _, id := range userIds {
			if id == uid {
				return true
			}
		}
		return false
	}
}

func FromSession(sessionId string) Middleware {
	return func(ctx *Context) bool {
		return ctx.Event.GetSessionId() == sessionId
	}
}

// 仅群组管理员
func FromAdmin() Middleware {
	return func(ctx *Context) bool {
		if ev, ok := ctx.Event.(*GroupMessageEvent); ok {
			return ev.Sender.Role == "admin"
		}
		return false
	}
}

// 群组管理员及更高权限，包括群主、超管
func FromAdminOrHigher() Middleware {
	admin := FromAdmin()
	owner := FromOwner()
	su := FromSuperuser()
	return func(ctx *Context) bool {
		return admin(ctx) || owner(ctx) || su(ctx)
	}
}

// 仅群组群主
func FromOwner() Middleware {
	return func(ctx *Context) bool {
		if ev, ok := ctx.Event.(*GroupMessageEvent); ok {
			return ev.Sender.Role == "owner"
		}
		return false
	}
}

// 群组群主及更高权限，包括超管
func FromOwnerOrHigher() Middleware {
	owner := FromOwner()
	su := FromSuperuser()
	return func(ctx *Context) bool {
		return owner(ctx) || su(ctx)
	}
}

// 仅超管，群聊和私聊都可
func FromSuperuser() Middleware {
	return func(ctx *Context) bool {
		config := ctx.Engine.Config
		sus := config.GetBaseConfig().Superuser

		var senderId int64
		if ev, ok := ctx.Event.(*GroupMessageEvent); ok {
			senderId = ev.Sender.UserId
		} else if ev, ok := ctx.Event.(*PrivateMessageEvent); ok {
			senderId = ev.Sender.UserId
		} else {
			return false
		}

		for _, su := range sus {
			if su == senderId {
				return true
			}
		}
		return false
	}
}

type prefixMatchResult struct {
	Matched string // 匹配到的前缀
	Remain  string // 去除前缀后的剩下文本
	Raw     string // 原始文本
}

// 事件为MessageEvent，且消息以某个前缀开头
func StartsWith(prefix ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)", strings.Join(prefix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("prefix", &prefixMatchResult{
			Matched: find,
			Remain:  strings.TrimPrefix(msgText, find),
			Raw:     msgText,
		})

		return true
	}
}

// 获取StartsWith匹配结果
func (ctx *Context) GetPrefixMatchResult() *prefixMatchResult {
	if v, ok := ctx.Get("prefix"); ok {
		return v.(*prefixMatchResult)
	}
	return nil
}

type suffixMatchResult struct {
	Matched string // 匹配到的后缀
	Remain  string // 去除后缀后的剩下文本
	Raw     string // 原始文本
}

// 事件为MessageEvent，且消息以某个后缀结尾
func EndsWith(suffix ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)$", strings.Join(suffix, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("suffix", &suffixMatchResult{
			Matched: find,
			Remain:  strings.TrimSuffix(msgText, find),
			Raw:     msgText,
		})
		return true
	}
}

// 获取EndsWith的匹配结果
func (ctx *Context) GetSuffixMatchResult() *suffixMatchResult {
	if v, ok := ctx.Get("suffix"); ok {
		return v.(*suffixMatchResult)
	}
	return nil
}

type commandMatchResult struct {
	CmdPrefix string   // 命令前缀
	Command   string   // 匹配到的命令
	Args      []string // 命令参数，以空格分割
	Remain    string   // 去除命令后的剩下文本
	Raw       string   // 原始文本
}

// 事件为MessageEvent，且消息开头为指令
func Command(cmd ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		cmdPrefixs := ctx.Engine.Config.GetBaseConfig().CmdPrefix
		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)(%s)", strings.Join(cmdPrefixs, "|"), strings.Join(cmd, "|")))
		find := reg.FindStringSubmatch(msgText)
		if find == nil {
			return false
		}

		remain := strings.TrimPrefix(msgText, find[0])
		args := strings.Split(remain, " ")
		argsFiltered := make([]string, 0)
		for _, arg := range args {
			if arg != "" {
				argsFiltered = append(argsFiltered, arg)
			}
		}
		ctx.Set("command", &commandMatchResult{
			CmdPrefix: find[1],
			Command:   find[2],
			Args:      argsFiltered,
			Remain:    remain,
			Raw:       msgText,
		})

		return true
	}
}

// 获取Command的匹配结果
func (ctx *Context) GetCommandMatchResult() *commandMatchResult {
	if v, ok := ctx.Get("command"); ok {
		return v.(*commandMatchResult)
	}
	return nil
}

// 完全匹配
func FullMatch(text ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(text, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("fullMatch", text)

		return true
	}
}

// 事件为MessageEvent，且消息中包含其中某个关键词
func Keyword(keywords ...string) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		reg := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
		find := reg.FindString(msgText)
		if find == "" {
			return false
		}

		ctx.Set("keyword", find)

		return true
	}
}

// 获取Keyword的匹配结果
func (ctx *Context) GetKeywordMatchResult() string {
	if v, ok := ctx.Get("keyword"); ok {
		return v.(string)
	}
	return ""
}

type regexpMatchResult struct {
	matchGroup []string
	groupNames []string
}

// 捕获组数量
func (r regexpMatchResult) Len() int {
	return len(r.matchGroup)
}

// 获取第i个捕获组，从1开始。0为整个匹配结果
func (r regexpMatchResult) Get(idx int) string {
	if idx >= len(r.matchGroup) {
		return ""
	}
	return r.matchGroup[idx]
}

// 使用捕获组名称来获取捕获组，仅当正则表达式中使用了命名捕获组时有效
func (r regexpMatchResult) GetByName(name string) string {
	for i, n := range r.groupNames {
		if n == name {
			return r.matchGroup[i]
		}
	}
	return ""
}

// 事件为MessageEvent，且消息中存在子串满足正则表达式
func Regex(regex regexp.Regexp) Middleware {
	return func(ctx *Context) bool {
		e := ctx.Event
		if !e.IsMessageEvent() {
			return false
		}

		msgText := e.ExtractPlainText()
		find := regex.FindStringSubmatch(msgText)
		if find == nil {
			return false
		}

		ctx.Set("regex", &regexpMatchResult{
			matchGroup: find,
			groupNames: regex.SubexpNames(),
		})

		return true
	}
}

func (ctx *Context) GetRegexMatchResult() *regexpMatchResult {
	if v, ok := ctx.Get("regex"); ok {
		return v.(*regexpMatchResult)
	}
	return nil
}

// ShellLikeCommand解析错误时要采取的动作
type ParseFailedAction int

const (
	// 无法解析时，自动回复错误信息
	ParseFailedAction_AutoReply ParseFailedAction = iota
	// 无法解析时，跳过这个Handler
	ParseFailedAction_Skip
	// 无法解析时，允许进入这个Handler，交由用户处理
	ParseFailedAction_LetMeHandle
)

type parseResult struct {
	Parser  *goarg.Parser
	RawArgs []string    // 从字符串切割出来的原始参数
	Args    interface{} // 解析后的参数结构体指针
	Err     error       // 解析失败时的错误信息
}

// 用法
func (res *parseResult) GetUsage() string {
	usage := bytes.NewBuffer(nil)
	res.Parser.WriteUsage(usage)
	return usage.String()
}

// 帮助，包含用法、参数说明、子命令说明
func (res *parseResult) GetHelp() string {
	help := bytes.NewBuffer(nil)
	res.Parser.WriteHelp(help)
	return help.String()
}

// 定义了子命令的情况下，返回是否解析到了子命令。未定义子命令时，总是返回true
func (res *parseResult) HasSubcommand() bool {
	return res.Parser.Subcommand() != nil
}

// 获取子命令结构体指针。
// 如果定义了子命令，但没有解析到子命令，返回nil。
// 如果没有定义子命令，返回最顶层的结构体指针。
func (res *parseResult) GetSubcommand() interface{} {
	return res.Parser.Subcommand()
}

// 生成错误提示
func (res *parseResult) FormatErrorAndHelp(err error) string {
	return fmt.Sprintf("%s\n%s", err.Error(), res.GetHelp())
}

// 命令行命令。
//
// cmd为命令名，会受命令前缀影响，例如：命令前缀为"!"，命令为"test"，则为"!test"。
//
// args为命令参数，传入结构体，**注意，不会向该结构体写入数据**。
// 用法见https://github.com/alexflint/go-arg。
//
// whenFailed为解析失败时的处理方式，推荐使用ParseFailedAction_AutoReply
func ShellLikeCommand(cmd string, args interface{}, whenFailed ParseFailedAction) Middleware {
	onCmd := Command(cmd)

	return func(ctx *Context) bool {
		if !onCmd(ctx) {
			return false
		}

		remain := ctx.GetCommandMatchResult().Remain
		remain = strings.TrimSpace(remain)
		argSlice := strings.Split(remain, " ")
		argSliceFiltered := make([]string, 0, len(argSlice))
		for _, arg := range argSlice {
			if arg != "" {
				argSliceFiltered = append(argSliceFiltered, arg)
			}
		}

		// 防止并发修改，创建一个新的结构体来存储解析结果
		pArgsCopy := createUnderlyingStruct(args) // 指针
		parser, err := goarg.NewParser(goarg.Config{Program: cmd, IgnoreEnv: true}, pArgsCopy)
		if err == nil {
			err = parser.Parse(argSliceFiltered)
			// 直接处理help参数并返回，不需要用户自己处理
			if err == goarg.ErrHelp {
				usage := bytes.NewBuffer(nil)
				parser.WriteHelp(usage)
				ctx.Reply(usage.String())
				ctx.Abort()
				return true
			}
		}

		result := &parseResult{
			Parser:  parser,
			RawArgs: argSliceFiltered,
			Args:    pArgsCopy,
			Err:     err,
		}
		ctx.Set("sh_cmd", result)

		if err != nil {
			switch whenFailed {
			case ParseFailedAction_AutoReply:
				ctx.Reply(fmt.Sprintf("参数解析失败：%v\n%s", err, result.GetUsage()))
				ctx.Abort()
				return true

			case ParseFailedAction_Skip:
				return false

			case ParseFailedAction_LetMeHandle:
				return true
			}
		}
		return true
	}
}

// 获取ShellLikeCommand的解析结果
func (ctx *Context) GetShellLikeCommandResult() *parseResult {
	r, ok := ctx.Get("sh_cmd")
	if !ok {
		return nil
	}
	return r.(*parseResult)
}
