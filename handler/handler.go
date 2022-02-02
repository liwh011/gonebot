package handler

// 处理函数
type HandleFunc func(*Context)

// 事件处理器
type EventHandler struct {
	handler       HandleFunc  // 处理函数
	preconditions []Condition // 运行的前置条件
	priority      int         // 优先级
	block         bool        // 是否阻塞
}

func (h *EventHandler) Handle(ctx *Context) {
	for _, precondition := range h.preconditions {
		if !precondition(ctx) {
			return
		}
	}
	h.handler(ctx)
}

// 设置优先级，从1开始，数字越小优先级越高。特例：-1为最低优先级
func (h *EventHandler) SetPriority(priority int) *EventHandler {
	h.priority = priority
	return h
}

// 设置是否阻塞，阻塞后，后续的handler不会被执行
func (h *EventHandler) SetBlock(block bool) *EventHandler {
	h.block = block
	return h
}

// 设置事件处理函数
func (h *EventHandler) SetHandler(handler HandleFunc) *EventHandler {
	h.handler = handler
	return h
}

// 设置该Handler的运行条件，仅当所有条件都满足时才会执行
func (h *EventHandler) AddPrecondition(precondition ...Condition) *EventHandler {
	h.preconditions = append(h.preconditions, precondition...)
	return h
}
