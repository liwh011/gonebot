package message

import (
	"encoding/json"
	"fmt"
)

type msgSegData map[string]interface{}

type MessageSegment struct {
	Type string     `json:"type"`
	Data msgSegData `json:"data"`
}

func (seg *MessageSegment) IsText() bool {
	return seg.Type == "text"
}

type Message struct {
	segs []MessageSegment
}

func (m *Message) GetSegment(index int) MessageSegment {
	return m.segs[index]
}

func (m *Message) Len() int {
	return len(m.segs)
}

type t_StringOrSegment interface{}

// 添加一个消息段（或字符串）到消息数组末尾。泛型版本的AppendXXXX
func (m *Message) Append(t t_StringOrSegment) (err error) {
	switch t := t.(type) {
	case string:
		m.AppendText(t)
	case MessageSegment:
		m.AppendSegment(t)
	default:
		err = fmt.Errorf("unknown type: %T", t)
	}
	return
}

// 添加一个消息段到消息数组末尾
func (m *Message) AppendSegment(seg MessageSegment) {
	m.segs = append(m.segs, seg)
}

// 添加一个字符串到消息数组末尾
func (m *Message) AppendText(t string) {
	m.AppendSegment(Text(t))
}

type t_MessageOrSegmentArray interface{}

// 拼接一个消息数组或多个消息段到消息数组末尾。泛型版本的ExtendXXXX
func (m *Message) Extend(msg t_MessageOrSegmentArray) (err error) {
	switch msg := msg.(type) {
	case Message:
		m.ExtendMessage(msg)
	case []MessageSegment:
		m.ExtendSegmentArray(msg)
	default:
		err = fmt.Errorf("unknown type: %T", msg)
	}
	return
}

// 拼接一个消息数组到消息数组末尾
func (m *Message) ExtendMessage(msg Message) {
	m.segs = append(m.segs, msg.segs...)
}

// 拼接多个消息段到消息数组末尾
func (m *Message) ExtendSegmentArray(segs []MessageSegment) {
	m.segs = append(m.segs, segs...)
}

// 提取消息内纯文本消息
func (m *Message) ExtractPlainText() (text string) {
	for _, seg := range m.segs {
		if seg.IsText() {
			text += seg.Data["text"].(string)
		}
	}
	return
}

func (m *Message) FilterByType(t string) []MessageSegment {
	segs := make([]MessageSegment, 0)
	for _, seg := range m.segs {
		if seg.Type == t {
			segs = append(segs, seg)
		}
	}
	return segs
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.segs)
}

func (m *Message) UnmarshalJSON(data []byte) (err error) {
	err = json.Unmarshal(data, &m.segs)
	return
}

type t_StringOrMsgOrSegOrArray interface{}

func Join(msgs ...t_StringOrMsgOrSegOrArray) (msg Message, err error) {
	msg = Message{}
	for _, m := range msgs {
		switch m := m.(type) {
		case string:
			msg.AppendText(m)

		case Message:
			msg.ExtendMessage(m)

		case MessageSegment:
			msg.AppendSegment(m)

		case []MessageSegment:
			msg.ExtendSegmentArray(m)

		case []string:
			for _, s := range m {
				msg.AppendText(s)
			}

		case []Message:
			for _, m := range m {
				msg.ExtendMessage(m)
			}

		case []interface{}:
			for _, m := range m {
				err = msg.Append(m)
			}

		default:
			err = fmt.Errorf("unknown type: %T", m)
		}
	}

	return
}

func Format(tmpl string, args ...interface{}) (msg Message, err error) {
	argsNoSeg := make([]interface{}, 0, len(args))
	for _, arg := range args {
		switch arg.(type) {
		case Message:
			continue
		case MessageSegment:
			continue
		default:
			argsNoSeg = append(argsNoSeg, arg)
		}
	}

	formattedTemplate := fmt.Sprintf(tmpl, argsNoSeg...)
	i := 0
	for i < len(formattedTemplate) {
		cur := formattedTemplate[i]
		if cur == '{' {
			if i+1 >= len(formattedTemplate) {
				break
			}
			next := formattedTemplate[i+1]
			if next == '{' {
				i += 2
				continue
			}
			if next == '}' {
				i += 2
				continue
			}
		}
		i += 1
	}

	return
}
