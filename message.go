package gonebot

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type Message []MessageSegment

func (m Message) Len() int {
	return len(m)
}

func (m Message) String() string {
	var s string
	for _, seg := range m {
		s += seg.String()
	}
	return s
}

type t_StringOrSegment interface{}

// 添加一个消息段（或字符串）到消息数组末尾。泛型版本的AppendXXXX
func (m *Message) Append(t t_StringOrSegment) {
	switch t := t.(type) {
	case string:
		m.AppendText(t)
	case MessageSegment:
		m.AppendSegment(t)
	default:
		m.AppendText(fmt.Sprintf("%v", t))
	}
}

// 添加一个消息段到消息数组末尾
func (m *Message) AppendSegment(seg MessageSegment) {
	*m = append(*m, seg)
}

// 添加一个字符串到消息数组末尾
func (m *Message) AppendText(t string) {
	m.AppendSegment(MsgFactory.Text(t))
}

type t_MessageOrSegmentArray interface{}

// 拼接一个消息数组或多个消息段到消息数组末尾。泛型版本的ExtendXXXX
func (m *Message) Extend(msg t_MessageOrSegmentArray) (err error) {
	switch msg := msg.(type) {
	case Message:
		m.ExtendMessage(msg)
	case *Message:
		m.ExtendMessage(*msg)
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
	text = strings.TrimSpace(text)
	return
}

func (m Message) FilterByType(segmentType string) []MessageSegment {
	segs := make([]MessageSegment, 0)
	for _, seg := range m {
		if seg.Type == segmentType {
			segs = append(segs, seg)
		}
	}
	return segs
}

// type t_StringOrMsgOrSegOrArray interface{}

// 将这些参数转换成一个Message
//
// 参数类型限制为string、(*)Message、MessageSegment、[]MessageSegment，
// 非上述类型的参数将被转换成一个Text消息段
func MsgPrint(msgs ...interface{}) (msg Message) {
	msg = Message{}
	for _, m := range msgs {
		switch m := m.(type) {
		case string:
			msg.AppendText(m)

		case Message:
			msg.ExtendMessage(m)

		case *Message:
			msg.ExtendMessage(*m)

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
				msg.Append(m)
			}

		default:
			msg.AppendText(fmt.Sprintf("%v", m))
		}
	}

	return
}

// 根据模板和参数生成消息对象。参数过多过少都会返回error。
//
// 使用"{}"作为消息段的占位符。例如：
//	Format("{}你好啊%s", At(114514), "李田所")
//	// 返回：[{at:114514},{text:"你好啊李田所"}]
// 如想"{}"不被解析为占位符，则使用"{{}}"。例如：
//	Format("{{}}你好啊%s", "李田所")
//	// 返回：[{text:"{}你好啊李田所"}]
func MsgPrintf(tmpl string, args ...interface{}) (msg Message, err error) {
	// 将普通参数、消息段参数分离开来
	argsNoSeg := make([]interface{}, 0, len(args))
	argsToFormat := make([]interface{}, 0, len(args))
	for _, arg := range args {
		switch arg := arg.(type) {
		case Message:
			argsToFormat = append(argsToFormat, arg)
		case *Message:
			argsToFormat = append(argsToFormat, arg)
		case MessageSegment:
			argsToFormat = append(argsToFormat, arg)
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
			// 如果是"{{}}"，则认为是{}
			if j+4 <= len(formattedTemplate) && formattedTemplate[j:j+4] == "{{}}" {
				builder.WriteString("{}")
				j += 4
				continue
			}

			// 遇到一对大括号，则将大括号前面的字符串生成一个Text消息段
			if j+1 <= len(formattedTemplate) && formattedTemplate[j+1] == '}' {
				if i != j {
					msg.AppendText(builder.String())
					builder.Reset()
				}
				if count >= len(argsToFormat) {
					err = fmt.Errorf("too few arguments for template: %s", tmpl)
					return
				}

				switch arg := argsToFormat[count].(type) {
				case Message:
					msg.ExtendMessage(arg)
				case *Message:
					msg.ExtendMessage(*arg)
				case MessageSegment:
					msg.AppendSegment(arg)
				}

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

	if count != len(argsToFormat) {
		err = fmt.Errorf("too many arguments for template: %s", tmpl)
	}

	return
}

// 作用同MsgPrintf，参数过多过少时会panic
func MsgMustPrintf(tmpl string, args ...interface{}) (msg Message) {
	msg, err := MsgPrintf(tmpl, args...)
	if err != nil {
		panic(err)
	}
	return
}

func convertJsonObjectToMessageSegment(m gjson.Result) (seg MessageSegment) {
	seg.Type = m.Get("type").String()
	if m.Get("data").Exists() {
		seg.Data = m.Get("data").Value().(map[string]interface{})
	}
	return
}

func convertJsonArrayToMessage(m []gjson.Result) (msg Message) {
	for _, m := range m {
		seg := convertJsonObjectToMessageSegment(m)
		msg.AppendSegment(seg)
	}
	return
}
