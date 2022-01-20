package message

import (
	"fmt"
	"strings"
)

type msgSegData map[string]interface{}

type MessageSegment struct {
	Type string     `json:"type"`
	Data msgSegData `json:"data"`
}

func (seg *MessageSegment) IsText() bool {
	return seg.Type == "text"
}

type Message []MessageSegment

func (m Message) Len() int {
	return len(m)
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
	*m = append(*m, seg)
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
	*m = append(*m, msg...)
}

// 拼接多个消息段到消息数组末尾
func (m *Message) ExtendSegmentArray(segs []MessageSegment) {
	*m = append(*m, segs...)
}

// 提取消息内纯文本消息
func (m Message) ExtractPlainText() (text string) {
	for _, seg := range m {
		if seg.IsText() {
			text += seg.Data["text"].(string)
		}
	}
	return
}

func (m Message) FilterByType(t string) []MessageSegment {
	segs := make([]MessageSegment, 0)
	for _, seg := range m {
		if seg.Type == t {
			segs = append(segs, seg)
		}
	}
	return segs
}

// func (m Message) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(m.segs)
// }

// func (m *Message) UnmarshalJSON(data []byte) (err error) {
// 	err = json.Unmarshal(data, &m.segs)
// 	return
// }

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

// 根据模板和参数生成消息对象。使用"{}"作为消息段的占位符。例如：
//	Format("{}你好啊%s", At(114514), "李田所")
//	// 返回：[{at:114514},{text:"你好啊李田所"}]
// 如想"{}"不被解析为占位符，则使用"{{}}"。例如：
//	Format("{{}}你好啊%s", "李田所")
//	// 返回：[{text:"{}你好啊李田所"}]
func Format(tmpl string, args ...interface{}) (msg Message, err error) {
	// 将普通参数、消息段参数分离开来
	argsNoSeg := make([]interface{}, 0, len(args))
	argsSeg := make([]MessageSegment, 0, len(args))
	for _, arg := range args {
		switch arg := arg.(type) {
		case Message:
			continue
		case MessageSegment:
			argsSeg = append(argsSeg, arg)
		default:
			argsNoSeg = append(argsNoSeg, arg)
		}
	}

	// 将"%xx"的占位符交给Sprintf来格式化，得到的字符串只剩下"{}"占位符
	formattedTemplate := fmt.Sprintf(tmpl, argsNoSeg...)

	// 将"{}"占位符替换成消息段。
	builder := strings.Builder{}
	builder.Grow(len(formattedTemplate))
	i := 0
	count := 0
	for j := i; j < len(formattedTemplate); {
		if formattedTemplate[j] == '{' {
			if j+4 <= len(formattedTemplate) && formattedTemplate[j:j+4] == "{{}}" {
				// 如果是"{{}}"，则认为是{}
				builder.WriteString("{}")
				j += 4
				continue
			}
			// if formattedTemplate[j+1] == '{' {
			// 	j += 2
			// 	continue
			// }
			if j+1 <= len(formattedTemplate) && formattedTemplate[j+1] == '}' {
				if i != j {
					msg.AppendText(builder.String())
					builder.Reset()
				}
				if count >= len(argsSeg) {
					err = fmt.Errorf("too few arguments for template: %s", tmpl)
					return
				}
				msg.AppendSegment(argsSeg[count])
				j += 2
				i = j
				count++
				continue
			}
		}
		builder.WriteByte(formattedTemplate[j])
		j++
	}
	if i < len(formattedTemplate) {
		// msg.AppendText(formattedTemplate[i:])
		msg.AppendText(builder.String())
		builder.Reset()
	}

	if count != len(argsSeg) {
		err = fmt.Errorf("too many arguments for template: %s", tmpl)
	}

	return
}
