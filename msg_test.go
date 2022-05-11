package gonebot

import (
	"fmt"
	"testing"
)

func Test_Image(t *testing.T) {

	a := MsgFactory.Image("http://www.baidu.com/img/bd_logo1.png",
		MsgFactory.ImageOptions().SetCache(false).SetProxy(true))
	fmt.Printf("%v\n", a)
}

func Test_Format(t *testing.T) {
	testData := []struct {
		format string
		args   []interface{}
		expect func(msg Message) bool
	}{
		{
			"%s",
			[]interface{}{"hello"},
			func(msg Message) bool {
				return len(msg) == 1 && msg[0].IsText() && msg[0].Data["text"] == "hello"
			},
		},
		{
			"%s %s %d",
			[]interface{}{"hello", "world", 114},
			func(msg Message) bool {
				return len(msg) == 1 && msg[0].IsText() && msg[0].Data["text"] == "hello world 114"
			},
		},
		{
			"word:%s, atsb:{}, num:%d",
			[]interface{}{"hello", MsgFactory.AtSomeone(114514), 114},
			func(msg Message) bool {
				return (len(msg) == 3 && msg[0].IsText() && msg[0].Data["text"] == "word:hello, atsb:" &&
					msg[1].Type == "at" && msg[2].IsText() && msg[2].Data["text"] == ", num:114")
			},
		},
		{
			"atsb:{}, num:%d, face:{}",
			[]interface{}{MsgFactory.AtSomeone(114514), 114, MsgFactory.Face(1919)},
			func(msg Message) bool {
				return (len(msg) == 4 &&
					msg[0].IsText() && msg[0].Data["text"] == "atsb:" &&
					msg[1].Type == "at" &&
					msg[2].IsText() && msg[2].Data["text"] == ", num:114, face:" &&
					msg[3].Type == "face")
			},
		},
		{
			"{}, num:%d, face:{} asdsa",
			[]interface{}{MsgFactory.AtSomeone(114514), 114, MsgFactory.Face(1919)},
			func(msg Message) bool {
				return (len(msg) == 4 &&
					msg[0].Type == "at" &&
					msg[1].IsText() && msg[1].Data["text"] == ", num:114, face:" &&
					msg[2].Type == "face" &&
					msg[3].IsText() && msg[3].Data["text"] == " asdsa")
			},
		},
		{
			"aa{{}}}",
			[]interface{}{},
			func(msg Message) bool {
				return (len(msg) == 1 && msg[0].IsText() && msg[0].Data["text"] == "aa{}}")
			},
		},
		{
			"aa{}",
			[]interface{}{MsgPrint("114514", MsgFactory.AtAll())},
			func(msg Message) bool {
				t.Logf("%v", msg)
				return true
			},
		},
	}

	for _, data := range testData {
		msg, err := MsgPrintf(data.format, data.args...)
		if err != nil {
			t.Errorf("Format(%q, %v) error: %v", data.format, data.args, err)
		}
		if !data.expect(msg) {
			t.Errorf("Format(%q, %v) = %v", data.format, data.args, msg)
		}
	}
}

func Test_String(t *testing.T) {
	testData := []struct {
		msg    Message
		expect string
	}{
		{
			Message{MsgFactory.AtAll(), MsgFactory.Text("hello")},
			"[CQ:at,qq=all]hello",
		},
		{
			Message{
				MsgFactory.Image("http://www.baidu.com/img/bd_logo1.png",
					MsgFactory.ImageOptions().SetCache(false).SetProxy(true)),
				MsgFactory.Text("hello"),
			},
			"[CQ:image,cache=0,proxy=1,file=http://www.baidu.com/img/bd_logo1.png]hello",
		},
		{
			Message{MsgFactory.Shake()},
			"[CQ:shake]",
		},
	}

	for i, data := range testData {
		if fmt.Sprintf("%v", data.msg) != data.expect {
			t.Errorf("%d: %v != %v", i, fmt.Sprintf("%v", data.msg), data.expect)
		}
	}
}
