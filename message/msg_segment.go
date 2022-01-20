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

type Message struct {
	segs []MessageSegment
}

func (m *Message) GetSegment(index int) MessageSegment {
	return m.segs[index]
}

func (m *Message) Len() int {
	return len(m.segs)
}

func (m *Message) Add(t interface{}) (err error) {
	switch t := t.(type) {
	case string:
		m.AddText(t)
	case Message:
		m.AddMessage(t)
	case MessageSegment:
		m.AddSegment(t)
	default:
		err = fmt.Errorf("unknown type: %T", t)
	}
	return
}

func (m *Message) AddSegment(seg MessageSegment) {
	m.segs = append(m.segs, seg)
}

func (m *Message) AddText(t string) {
	m.AddSegment(Text(t))
}

func (m *Message) AddMessage(msg Message) {
	m.segs = append(m.segs, msg.segs...)
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.segs)
}
