package gonebot

import (
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type ApiParams map[string]interface{}

func (bot *Bot) CallApi(action string, params ApiParams) (*gjson.Result, error) {
	log.Infof("正在调用接口%s", action)
	rsp, err := bot.adapter.Request(action, params)
	if err != nil {
		log.Errorf("调用接口%s失败: %s", action, err)
		return nil, err
	}
	data := rsp.(response).Data
	return &data, nil
}

// 发送私聊消息
func (bot *Bot) SendPrivateMsg(userId int64, message Message, autoEscape bool) (int32, error) {
	data, err := bot.CallApi("send_private_msg", ApiParams{
		"user_id":     userId,
		"message":     message.String(),
		"auto_escape": autoEscape,
	})
	if err != nil {
		return -1, err
	}
	messageId := int32(data.Get("message_id").Int())
	return messageId, nil
}

// 发送群消息
func (bot *Bot) SendGroupMsg(groupId int64, message Message, autoEscape bool) (int32, error) {
	data, err := bot.CallApi("send_group_msg", ApiParams{
		"group_id":    groupId,
		"message":     message.String(),
		"auto_escape": autoEscape,
	})
	if err != nil {
		return -1, err
	}
	messageId := int32(data.Get("message_id").Int())
	return messageId, nil
}

// 发送消息
func (bot *Bot) SendMsg(messageType string, userId, groupId int64, message Message, autoEscape bool) (int32, error) {
	params := ApiParams{
		"message_type": messageType,
		"message":      message.String(),
		"auto_escape":  autoEscape,
	}
	switch messageType {
	case "private":
		params["user_id"] = userId
	case "group":
		params["group_id"] = groupId
	default: // 纠正错误的消息类型
		if groupId > 0 {
			params["message_type"] = "group"
		} else if userId > 0 {
			params["message_type"] = "private"
		} else {
			return -1, ErrInvalidMessageType
		}
	}
	data, err := bot.CallApi("send_msg", params)
	if err != nil {
		return -1, err
	}
	messageId := int32(data.Get("message_id").Int())
	return messageId, nil
}

// 撤回消息
func (bot *Bot) DeleteMsg(messageId int32) error {
	_, err := bot.CallApi("delete_msg", ApiParams{
		"message_id": messageId,
	})
	return err
}

type ApiResponse_GetMsg struct {
	Time          int32
	MessageType   string
	MessageId     int32
	RealId        int32
	privateSender *MessageEventSender
	groupSender   *GroupMessageEventSender
	Message       Message
}

func (r *ApiResponse_GetMsg) GetPrivateSender() *MessageEventSender {
	return r.privateSender
}

func (r *ApiResponse_GetMsg) GetGroupSender() *GroupMessageEventSender {
	return r.groupSender
}

// 获取消息
func (bot *Bot) GetMsg(messageId int32) (*ApiResponse_GetMsg, error) {
	data, err := bot.CallApi("get_msg", ApiParams{
		"message_id": messageId,
	})
	if err != nil {
		return nil, err
	}

	rsp := &ApiResponse_GetMsg{
		Time:        int32(data.Get("time").Int()),
		MessageType: data.Get("message_type").String(),
		MessageId:   int32(data.Get("message_id").Int()),
		RealId:      int32(data.Get("real_id").Int()),
		Message:     ConvertJsonArrayToMessage(data.Get("message").Array()),
	}
	sender := MessageEventSender{
		UserId:   data.Get("user_id").Int(),
		Nickname: data.Get("nickname").String(),
		Sex:      data.Get("sex").String(),
		Age:      int32(data.Get("age").Int()),
	}
	switch rsp.MessageType {
	case "private":
		rsp.privateSender = &sender
	case "group":
		gsender := GroupMessageEventSender{
			MessageEventSender: sender,
			Card:               data.Get("card").String(),
			Area:               data.Get("area").String(),
			Level:              data.Get("level").String(),
			Role:               data.Get("role").String(),
			Title:              data.Get("title").String(),
		}
		rsp.groupSender = &gsender
	default:
		return nil, ErrInvalidMessageType
	}
	return rsp, nil
}

// 获取合并转发消息
func (bot *Bot) GetForwardMsg(messageId int32) (*Message, error) {
	data, err := bot.CallApi("get_forward_msg", ApiParams{
		"message_id": messageId,
	})
	if err != nil {
		return nil, err
	}
	ret := ConvertJsonArrayToMessage(data.Get("message").Array())
	return &ret, nil
}

// 发送好友赞
func (bot *Bot) SendLike(userId int64, times int) error {
	_, err := bot.CallApi("send_like", ApiParams{
		"user_id": userId,
		"times":   times,
	})
	return err
}

// 群组踢人
func (bot *Bot) SetGroupKick(groupId int64, userId int64, rejectAddRequest bool) error {
	_, err := bot.CallApi("set_group_kick", ApiParams{
		"group_id":           groupId,
		"user_id":            userId,
		"reject_add_request": rejectAddRequest,
	})
	return err
}

// 群组单人禁言
func (bot *Bot) SetGroupBan(groupId int64, userId int64, duration int) error {
	_, err := bot.CallApi("set_group_ban", ApiParams{
		"group_id": groupId,
		"user_id":  userId,
		"duration": duration,
	})
	return err
}

// 群组匿名用户禁言
func (bot *Bot) SetGroupAnonymousBan(groupId int64, anonymous *Anonymous, anonymousFlag string, duration int) error {
	params := ApiParams{
		"group_id": groupId,
		"duration": duration,
	}
	if anonymous != nil {
		params["anonymous"] = anonymous
	} else if anonymousFlag != "" {
		params["anonymous_flag"] = anonymousFlag
	} else {
		return ErrMissingAnonymous
	}
	_, err := bot.CallApi("set_group_anonymous_ban", params)
	return err
}

// 群组全员禁言
func (bot *Bot) SetGroupWholeBan(groupId int64, enable bool) error {
	_, err := bot.CallApi("set_group_whole_ban", ApiParams{
		"group_id": groupId,
		"enable":   enable,
	})
	return err
}

// 群组设置管理员
func (bot *Bot) SetGroupAdmin(groupId int64, userId int64, enable bool) error {
	_, err := bot.CallApi("set_group_admin", ApiParams{
		"group_id": groupId,
		"user_id":  userId,
		"enable":   enable,
	})
	return err
}

// 群组开关匿名
func (bot *Bot) SetGroupAnonymous(groupId int64, enable bool) error {
	_, err := bot.CallApi("set_group_anonymous", ApiParams{
		"group_id": groupId,
		"enable":   enable,
	})
	return err
}

// 设置群名片（群备注）
func (bot *Bot) SetGroupCard(groupId int64, userId int64, card string) error {
	_, err := bot.CallApi("set_group_card", ApiParams{
		"group_id": groupId,
		"user_id":  userId,
		"card":     card,
	})
	return err
}

// 设置群名
func (bot *Bot) SetGroupName(groupId int64, groupName string) error {
	_, err := bot.CallApi("set_group_name", ApiParams{
		"group_id":   groupId,
		"group_name": groupName,
	})
	return err
}

// 退出群组
func (bot *Bot) SetGroupLeave(groupId int64) error {
	_, err := bot.CallApi("set_group_leave", ApiParams{
		"group_id": groupId,
	})
	return err
}

// 解散群组
func (bot *Bot) SetGroupDismiss(groupId int64) error {
	_, err := bot.CallApi("set_group_leave", ApiParams{
		"group_id":   groupId,
		"is_dismiss": true,
	})
	return err
}

// 设置群组专属头衔
func (bot *Bot) SetGroupSpecialTitle(groupId int64, userId int64, specialTitle string, duration int) error {
	_, err := bot.CallApi("set_group_special_title", ApiParams{
		"group_id":      groupId,
		"user_id":       userId,
		"special_title": specialTitle,
		"duration":      duration,
	})
	return err
}

// 处理加好友请求
func (bot *Bot) SetFriendAddRequest(flag string, approve bool, remark string) error {
	_, err := bot.CallApi("set_friend_add_request", ApiParams{
		"flag":    flag,
		"approve": approve,
		"remark":  remark,
	})
	return err
}

// 处理加群请求／邀请
func (bot *Bot) SetGroupAddRequest(flag string, subType int, approve bool, reason string) error {
	_, err := bot.CallApi("set_group_add_request", ApiParams{
		"flag":     flag,
		"sub_type": subType,
		"approve":  approve,
		"reason":   reason,
	})
	return err
}

type LoginInfo struct {
	UserId   int64
	Nickname string
}

// 获取登录号信息
func (bot *Bot) GetLoginInfo() (*LoginInfo, error) {
	resp, err := bot.CallApi("get_login_info", nil)
	if err != nil {
		return nil, err
	}
	info := LoginInfo{
		UserId:   resp.Get("user_id").Int(),
		Nickname: resp.Get("nickname").String(),
	}
	return &info, nil
}

type StrangerInfo struct {
	UserId   int64
	Nickname string
	Sex      string
	Age      int32
}

// 获取陌生人信息
func (bot *Bot) GetStrangerInfo(userId int64) (*StrangerInfo, error) {
	resp, err := bot.CallApi("get_stranger_info", ApiParams{
		"user_id": userId,
	})
	if err != nil {
		return nil, err
	}
	info := StrangerInfo{
		UserId:   resp.Get("user_id").Int(),
		Nickname: resp.Get("nickname").String(),
		Sex:      resp.Get("sex").String(),
		Age:      int32(resp.Get("age").Int()),
	}
	return &info, nil
}

type FriendInfo struct {
	UserId   int64
	Nickname string
	Remark   string
}

// 获取好友列表
func (bot *Bot) GetFriendList() (*[]FriendInfo, error) {
	resp, err := bot.CallApi("get_friend_list", nil)
	if err != nil {
		return nil, err
	}
	var friends []FriendInfo
	for _, v := range resp.Get("friends").Array() {
		friend := FriendInfo{
			UserId:   v.Get("user_id").Int(),
			Nickname: v.Get("nickname").String(),
			Remark:   v.Get("remark").String(),
		}
		friends = append(friends, friend)
	}
	return &friends, nil
}

type GroupInfo struct {
	GroupId        int64
	GroupName      string
	MemberCount    int
	MaxMemberCount int
}

// 获取群信息
func (bot *Bot) GetGroupInfo(groupId int64) (*GroupInfo, error) {
	resp, err := bot.CallApi("get_group_info", ApiParams{
		"group_id": groupId,
	})
	if err != nil {
		return nil, err
	}
	info := GroupInfo{
		GroupId:        resp.Get("group_id").Int(),
		GroupName:      resp.Get("group_name").String(),
		MemberCount:    int(resp.Get("member_count").Int()),
		MaxMemberCount: int(resp.Get("max_member_count").Int()),
	}
	return &info, nil
}

// 获取群列表
func (bot *Bot) GetGroupList() (*[]GroupInfo, error) {
	resp, err := bot.CallApi("get_group_list", nil)
	if err != nil {
		return nil, err
	}
	var groups []GroupInfo
	for _, v := range resp.Array() {
		group := GroupInfo{
			GroupId:        v.Get("group_id").Int(),
			GroupName:      v.Get("group_name").String(),
			MemberCount:    int(v.Get("member_count").Int()),
			MaxMemberCount: int(v.Get("max_member_count").Int()),
		}
		groups = append(groups, group)
	}
	return &groups, nil
}

type GroupMemberInfo struct {
	GroupId         int64
	UserId          int64
	Nickname        string
	Card            string
	Sex             string
	Age             int
	Area            string
	JoinTime        int
	LastSentTime    int
	Level           string
	Role            string
	Unfriendly      bool
	Title           string
	TitleExpireTime int
	CardChangeable  bool
}

// 获取群成员信息
func (bot *Bot) GetGroupMemberInfo(groupId int64, userId int64) (*GroupMemberInfo, error) {
	resp, err := bot.CallApi("get_group_member_info", ApiParams{
		"group_id": groupId,
		"user_id":  userId,
	})
	if err != nil {
		return nil, err
	}
	info := GroupMemberInfo{
		GroupId:         resp.Get("group_id").Int(),
		UserId:          resp.Get("user_id").Int(),
		Nickname:        resp.Get("nickname").String(),
		Card:            resp.Get("card").String(),
		Sex:             resp.Get("sex").String(),
		Age:             int(resp.Get("age").Int()),
		Area:            resp.Get("area").String(),
		JoinTime:        int(resp.Get("join_time").Int()),
		LastSentTime:    int(resp.Get("last_sent_time").Int()),
		Level:           resp.Get("level").String(),
		Role:            resp.Get("role").String(),
		Unfriendly:      resp.Get("unfriendly").Bool(),
		Title:           resp.Get("title").String(),
		TitleExpireTime: int(resp.Get("title_expire_time").Int()),
		CardChangeable:  resp.Get("card_changeable").Bool(),
	}
	return &info, nil
}

// 获取群成员列表
func (bot *Bot) GetGroupMemberList(groupId int64) (*[]GroupMemberInfo, error) {
	resp, err := bot.CallApi("get_group_member_list", ApiParams{
		"group_id": groupId,
	})
	if err != nil {
		return nil, err
	}
	var members []GroupMemberInfo
	for _, v := range resp.Get("members").Array() {
		member := GroupMemberInfo{
			GroupId:         v.Get("group_id").Int(),
			UserId:          v.Get("user_id").Int(),
			Nickname:        v.Get("nickname").String(),
			Card:            v.Get("card").String(),
			Sex:             v.Get("sex").String(),
			Age:             int(v.Get("age").Int()),
			Area:            v.Get("area").String(),
			JoinTime:        int(v.Get("join_time").Int()),
			LastSentTime:    int(v.Get("last_sent_time").Int()),
			Level:           v.Get("level").String(),
			Role:            v.Get("role").String(),
			Unfriendly:      v.Get("unfriendly").Bool(),
			Title:           v.Get("title").String(),
			TitleExpireTime: int(v.Get("title_expire_time").Int()),
			CardChangeable:  v.Get("card_changeable").Bool(),
		}
		members = append(members, member)
	}
	return &members, nil
}

type CurrentTalkative struct {
	UserId   int64
	Nickname string
	Avatar   string
	DayCount int
}

type HonorListItem struct {
	UserId      int64
	Nickname    string
	Avatar      string
	Description string
}

type GroupHonorInfo struct {
	GroupId          int64
	CurrentTalkative *CurrentTalkative
	TalkativeList    []HonorListItem
	PerformerList    []HonorListItem
	LegendList       []HonorListItem
	StrongNewbieList []HonorListItem
	EmotionList      []HonorListItem
}

// 获取群荣誉信息
func (bot *Bot) GetGroupHonorInfo(groupId int64, type_ string) (*GroupHonorInfo, error) {
	resp, err := bot.CallApi("get_group_honor_info", ApiParams{
		"group_id": groupId,
		"type":     type_,
	})
	if err != nil {
		return nil, err
	}

	info := GroupHonorInfo{
		GroupId: resp.Get("group_id").Int(),
	}

	if type_ == "current_talkative" {
		talk := CurrentTalkative{
			UserId:   resp.Get("user_id").Int(),
			Nickname: resp.Get("nickname").String(),
			Avatar:   resp.Get("avatar").String(),
			DayCount: int(resp.Get("day_count").Int()),
		}
		info.CurrentTalkative = &talk
		return &info, nil
	}

	key := type_ + "_list"
	var honorList []HonorListItem
	for _, v := range resp.Get(key).Array() {
		item := HonorListItem{
			UserId:      v.Get("user_id").Int(),
			Nickname:    v.Get("nickname").String(),
			Avatar:      v.Get("avatar").String(),
			Description: v.Get("description").String(),
		}
		honorList = append(honorList, item)
	}
	switch type_ {
	case "talkative_list":
		info.TalkativeList = honorList
	case "performer_list":
		info.PerformerList = honorList
	case "legend_list":
		info.LegendList = honorList
	case "strong_newbie_list":
		info.StrongNewbieList = honorList
	case "emotion_list":
		info.EmotionList = honorList
	}
	return &info, nil
}

// 获取 Cookies
func (bot *Bot) GetCookies(domain string) (string, error) {
	resp, err := bot.CallApi("get_cookies", ApiParams{
		"domain": domain,
	})
	if err != nil {
		return "", err
	}
	return resp.Get("cookies").String(), nil
}

// 获取 CSRF Token
func (bot *Bot) GetCsrfToken() (int32, error) {
	resp, err := bot.CallApi("get_csrf_token", nil)
	if err != nil {
		return 0, err
	}
	return int32(resp.Get("csrf_token").Int()), nil
}

type QQCredential struct {
	Cookies   string
	CsrfToken int32
}

// 获取 QQ 相关接口凭证
func (bot *Bot) GetCredentials(domain string) (*QQCredential, error) {
	resp, err := bot.CallApi("get_credentials", nil)
	if err != nil {
		return nil, err
	}

	return &QQCredential{
		Cookies:   resp.Get("cookies").String(),
		CsrfToken: int32(resp.Get("csrf_token").Int()),
	}, nil
}

// 获取语音
func (bot *Bot) GetRecord(file string, outFormat string) (string, error) {
	resp, err := bot.CallApi("get_record", ApiParams{
		"file":       file,
		"out_format": outFormat,
	})
	if err != nil {
		return "", err
	}
	return resp.Get("file").String(), nil
}

// 获取图片
func (bot *Bot) GetImage(file string) (string, error) {
	resp, err := bot.CallApi("get_image", ApiParams{
		"file": file,
	})
	if err != nil {
		return "", err
	}
	return resp.Get("file").String(), nil
}

// 检查是否可以发送图片
func (bot *Bot) CanSendImage() (bool, error) {
	resp, err := bot.CallApi("can_send_image", nil)
	if err != nil {
		return false, err
	}
	return resp.Get("yes").Bool(), nil
}

// 检查是否可以发送语音
func (bot *Bot) CanSendRecord() (bool, error) {
	resp, err := bot.CallApi("can_send_record", nil)
	if err != nil {
		return false, err
	}
	return resp.Get("yes").Bool(), nil
}

type Status struct {
	Online bool
	Good   bool
}

// 获取运行状态
func (bot *Bot) GetStatus() (*Status, error) {
	resp, err := bot.CallApi("get_status", nil)
	if err != nil {
		return nil, err
	}
	return &Status{
		Online: resp.Get("online").Bool(),
		Good:   resp.Get("good").Bool(),
	}, nil
}

type VesionInfo struct {
	AppName         string
	AppVersion      string
	ProtocolVersion string
}

// 获取版本信息
func (bot *Bot) GetVersionInfo() (*VesionInfo, error) {
	resp, err := bot.CallApi("get_version_info", nil)
	if err != nil {
		return nil, err
	}
	return &VesionInfo{
		AppName:         resp.Get("app_name").String(),
		AppVersion:      resp.Get("app_version").String(),
		ProtocolVersion: resp.Get("protocol_version").String(),
	}, nil
}

// 重启 OneBot 实现
func (bot *Bot) SetRestart(delay int) error {
	_, err := bot.CallApi("set_restart", ApiParams{
		"delay": delay,
	})
	return err
}

// 清理缓存
func (bot *Bot) ClearCache() error {
	_, err := bot.CallApi("clear_cache", nil)
	return err
}

type quickOperationParams map[string]interface{}

// 对事件执行快速操作
func (bot *Bot) handleQuickOperation(context interface{}, operation quickOperationParams) error {
	_, err := bot.CallApi(".handle_quick_operation", ApiParams{
		"context":   context,
		"operation": operation,
	})
	return err
}
