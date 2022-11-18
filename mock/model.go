package mock

import (
	"fmt"
	"strings"
	"time"

	"github.com/liwh011/gonebot"
)

type User struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

type GroupMember struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"` // 昵称
	Age      int32  `json:"age"`
	Sex      string `json:"sex"`

	Card  string `json:"card"`  // 群名片、群备注
	Role  string `json:"role"`  // 角色 member owner admin
	Title string `json:"title"` // 头衔
	Level string `json:"level"` // 等级
	Area  string `json:"area"`  // 地区
}

// 设置为管理员
func (m *GroupMember) SetAdmin() {
	m.Role = "admin"
}

// 设置为群主
func (m *GroupMember) SetOwner() {
	m.Role = "owner"
}

// 设置为普通成员
func (m *GroupMember) SetMember() {
	m.Role = "member"
}

type Group struct {
	GroupId     int64         `json:"group_id"`     // 群号
	GroupName   string        `json:"group_name"`   // 群名
	MemberCount int32         `json:"member_count"` // 群成员数
	Members     []GroupMember `json:"-"`
}

func (group *Group) GetMember(userId int64) *GroupMember {
	for _, member := range group.Members {
		if member.UserId == userId {
			return &member
		}
	}
	return nil
}

func (group *Group) AddMember(member GroupMember) {
	if group.GetMember(member.UserId) == nil {
		group.Members = append(group.Members, member)
	}
	group.MemberCount = int32(len(group.Members))
}

func (group *Group) RemoveMember(userId int64) {
	for i, member := range group.Members {
		if member.UserId == userId {
			group.Members = append(group.Members[:i], group.Members[i+1:]...)
			return
		}
	}
	group.MemberCount = int32(len(group.Members))
}

// 一条消息
type MessageRecord struct {
	MsgId     int32
	Msg       gonebot.Message
	UserId    int64  // 发送者
	Nickname  string // 发送者昵称
	Time      int64  // 发送时间，unix时间戳
	GroupId   int64  // 群号，私聊为0
	SessionId int64  // 私聊时为对方的 QQ 号，群聊时为群号
}

func (r *MessageRecord) String() string {
	timeStr := time.Unix(r.Time, 0).Format("2006-01-02 15:04:05")
	if r.IsGroupMessage() {
		return fmt.Sprintf("[%s %s(%d)]\n%s", timeStr, r.Nickname, r.UserId, r.Msg)
	} else {
		return fmt.Sprintf("[%s %s]\n%s", timeStr, r.Nickname, r.Msg)
	}
}

func (r *MessageRecord) IsGroupMessage() bool {
	return r.GroupId != 0
}

type MessageHistory []MessageRecord

func (history MessageHistory) Len() int {
	return len(history)
}

func (history MessageHistory) Less(i, j int) bool {
	return history[i].Time < history[j].Time
}

func (history MessageHistory) Swap(i, j int) {
	history[i], history[j] = history[j], history[i]
}

func (history MessageHistory) getPrivateHistory(userId int64) MessageHistory {
	var result MessageHistory
	for _, record := range history {
		if !record.IsGroupMessage() && record.SessionId == userId {
			result = append(result, record)
		}
	}
	return result
}

func (history MessageHistory) getGroupHistory(groupId int64) MessageHistory {
	var result MessageHistory
	for _, record := range history {
		if record.GroupId == groupId {
			result = append(result, record)
		}
	}
	return result
}

func (history MessageHistory) String() string {
	ret := strings.Builder{}
	for _, record := range history {
		ret.WriteString(record.String())
		ret.WriteString("\n")
	}
	return ret.String()
}
