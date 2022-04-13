package gonebot

import (
	"fmt"
	"strconv"
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

func (seg MessageSegment) String() string {
	if seg.IsText() {
		return Escape(seg.Data["text"].(string), false)
	}

	if len(seg.Data) == 0 {
		return fmt.Sprintf("[CQ:%s]", seg.Type)
	}

	params := make([]string, 0, len(seg.Data))
	for k, v := range seg.Data {
		vStr := fmt.Sprintf("%v", v)
		params = append(params, fmt.Sprintf("%s=%s", k, Escape(vStr, false)))
	}
	return fmt.Sprintf("[CQ:%s,%s]", seg.Type, strings.Join(params, ","))

}

type messageSegmentFactory struct{}

var MessageSegmentFactory = messageSegmentFactory{}

// 纯文本
func (f messageSegmentFactory) Text(text string) MessageSegment {
	return MessageSegment{
		Type: "text",
		Data: msgSegData{
			"text": text,
		},
	}
}

// 表情。id为QQ表情ID
func (f messageSegmentFactory) Face(id int) MessageSegment {
	return MessageSegment{
		Type: "face",
		Data: msgSegData{
			"id": strconv.Itoa(id),
		},
	}
}

type imageOptions msgSegData

// 图片的可选参数
func (f messageSegmentFactory) ImageOptions() *imageOptions {
	return (&imageOptions{}).SetCache(true).SetProxy(true)
}

// 图片类型，"flash"表示闪照，无此参数表示普通图片
func (p *imageOptions) SetType(t string) *imageOptions {
	(*p)["type"] = t
	return p
}

// 只在通过网络URL发送时有效，表示是否使用已缓存的文件，默认true
func (p *imageOptions) SetCache(b bool) *imageOptions {
	(*p)["cache"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络URL发送时有效，表示是否通过代理下载文件（需通过环境变量或配置文件配置代理），默认true
func (p *imageOptions) SetProxy(b bool) *imageOptions {
	(*p)["proxy"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络URL发送时有效，单位秒，表示下载网络文件的超时时间，默认不超时
func (p *imageOptions) SetTimeout(t int) *imageOptions {
	(*p)["timeout"] = strconv.Itoa(t)
	return p
}

// 图片。file可以为网络URL、本地URI、Base64
func (f messageSegmentFactory) Image(file string, optional *imageOptions) MessageSegment {
	if optional == nil {
		optional = f.ImageOptions()
	}
	data := *optional
	data["file"] = file
	return MessageSegment{
		Type: "image",
		Data: msgSegData(data),
	}
}

type recordOptions msgSegData

// 语音的可选参数
func (f messageSegmentFactory) RecordOptions() *recordOptions {
	return (&recordOptions{}).SetCache(true).SetProxy(true).SetMagic(false)
}

// 默认 0，设置为 1 表示变声
func (p *recordOptions) SetMagic(b bool) *recordOptions {
	(*p)["magic"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络 URL 发送时有效，表示是否使用已缓存的文件，默认 1
func (p *recordOptions) SetCache(b bool) *recordOptions {
	(*p)["cache"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络 URL 发送时有效，表示是否通过代理下载文件（需通过环境变量或配置文件配置代理），默认 1
func (p *recordOptions) SetProxy(b bool) *recordOptions {
	(*p)["proxy"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络 URL 发送时有效，单位秒，表示下载网络文件的超时时间 ，默认不超时
func (p *recordOptions) SetTimeout(t int) *recordOptions {
	(*p)["timeout"] = strconv.Itoa(t)
	return p
}

// 语音
func (f messageSegmentFactory) Record(file string, optional *recordOptions) MessageSegment {
	if optional == nil {
		optional = f.RecordOptions()
	}
	data := msgSegData(*optional)
	data["file"] = file

	return MessageSegment{
		Type: "record",
		Data: data,
	}
}

type videoOptions msgSegData

// 视频的可选参数
func (f messageSegmentFactory) VideoOptions() *videoOptions {
	return (&videoOptions{}).SetCache(true).SetProxy(true)
}

// 只在通过网络 URL 发送时有效，表示是否使用已缓存的文件，默认 1
func (p *videoOptions) SetCache(b bool) *videoOptions {
	(*p)["cache"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络 URL 发送时有效，表示是否通过代理下载文件（需通过环境变量或配置文件配置代理），默认 1
func (p *videoOptions) SetProxy(b bool) *videoOptions {
	(*p)["proxy"] = strconv.Itoa(boolToInt01(b))
	return p
}

// 只在通过网络 URL 发送时有效，单位秒，表示下载网络文件的超时时间 ，默认不超时
func (p *videoOptions) SetTimeout(t int) *videoOptions {
	(*p)["timeout"] = strconv.Itoa(t)
	return p
}

// 视频
func (f messageSegmentFactory) Video(file string, optional *videoOptions) MessageSegment {
	if optional == nil {
		optional = f.VideoOptions()
	}
	data := msgSegData(*optional)
	data["file"] = file

	return MessageSegment{
		Type: "video",
		Data: data,
	}
}

// 群聊At指定QQ
func (f messageSegmentFactory) AtSomeone(qq int64) MessageSegment {
	return MessageSegment{
		Type: "at",
		Data: msgSegData{
			"qq": strconv.FormatInt(qq, 10),
		},
	}
}

// 群聊At全体
func (f messageSegmentFactory) AtAll() MessageSegment {
	return MessageSegment{
		Type: "at",
		Data: msgSegData{
			"qq": "all",
		},
	}
}

// 猜拳魔法表情
func (f messageSegmentFactory) Rps() MessageSegment {
	return MessageSegment{
		Type: "rps",
		Data: msgSegData{},
	}
}

// 掷骰子魔法表情
func (f messageSegmentFactory) Dice() MessageSegment {
	return MessageSegment{
		Type: "dice",
		Data: msgSegData{},
	}
}

// 窗口抖动（戳一戳）
func (f messageSegmentFactory) Shake() MessageSegment {
	return MessageSegment{
		Type: "shake",
		Data: msgSegData{},
	}
}

// 戳一戳
func (f messageSegmentFactory) Poke(type_, id int) MessageSegment {
	return MessageSegment{
		Type: "poke",
		Data: msgSegData{
			"type": strconv.Itoa(type_),
			"id":   strconv.Itoa(id),
		},
	}
}

// 匿名发消息
func (f messageSegmentFactory) AnonymousSegment(ignore bool) MessageSegment {
	return MessageSegment{
		Type: "anonymous",
		Data: msgSegData{
			"ignore": strconv.Itoa(boolToInt01(ignore)),
		},
	}
}

type shareOptions msgSegData

func (f messageSegmentFactory) ShareOptions() *shareOptions {
	return &shareOptions{}
}

// 可选，内容描述
func (p *shareOptions) SetContent(c string) *shareOptions {
	(*p)["content"] = c
	return p
}

// 可选，图片 URL
func (p *imageOptions) SetImage(image string) *imageOptions {
	(*p)["image"] = image
	return p
}

// 链接分享
func (f messageSegmentFactory) Share(url, title string, optional *shareOptions) MessageSegment {
	if optional == nil {
		optional = f.ShareOptions()
	}
	data := msgSegData(*optional)
	data["url"] = url
	data["title"] = title

	return MessageSegment{
		Type: "share",
		Data: data,
	}
}

// 推荐好友
func (f messageSegmentFactory) ContactQQ(id int64) MessageSegment {
	return MessageSegment{
		Type: "contact",
		Data: msgSegData{
			"type": "qq",
			"id":   strconv.FormatInt(id, 10),
		},
	}
}

// 推荐群
func (f messageSegmentFactory) ContactGroup(id int64) MessageSegment {
	return MessageSegment{
		Type: "contact",
		Data: msgSegData{
			"type": "group",
			"id":   strconv.FormatInt(id, 10),
		},
	}
}

type locationOptions msgSegData

func (f messageSegmentFactory) LocationOptions() *locationOptions {
	return &locationOptions{}
}

// 可选，标题
func (p *locationOptions) SetTitle(title string) *locationOptions {
	(*p)["title"] = title
	return p
}

// 可选，内容描述
func (p *locationOptions) SetContent(content string) *locationOptions {
	(*p)["content"] = content
	return p
}

// 位置
func (f messageSegmentFactory) Location(lat, lng float64, options *locationOptions) MessageSegment {
	if options == nil {
		options = f.LocationOptions()
	}
	data := msgSegData(*options)
	data["lat"] = strconv.FormatFloat(lat, 'f', -1, 64)
	data["lng"] = strconv.FormatFloat(lng, 'f', -1, 64)

	return MessageSegment{
		Type: "location",
		Data: data,
	}
}

const (
	MUSIC_TYPE_QQ    = "qq"  // QQ 音乐
	MUSIC_TYPE_163   = "163" // 网易云音乐
	MUSIC_TYPE_XIAMI = "xm"  // 虾米音乐
)

// 音乐分享
func (f messageSegmentFactory) Music(id, type_ string) MessageSegment {
	return MessageSegment{
		Type: "music",
		Data: msgSegData{
			"type": type_,
			"id":   id,
		},
	}
}

type customMusicParams msgSegData

func (f messageSegmentFactory) CustomMusicParams() *customMusicParams {
	return &customMusicParams{}
}

// 可选，内容描述
func (p *customMusicParams) SetContent(c string) *customMusicParams {
	(*p)["content"] = c
	return p
}

// 可选，图片 URL
func (p *customMusicParams) SetImage(image string) *customMusicParams {
	(*p)["image"] = image
	return p
}

// 音乐自定义分享
func (f messageSegmentFactory) CustomMusic(url, title, audioUrl string, optional *customMusicParams) MessageSegment {
	if optional == nil {
		optional = f.CustomMusicParams()
	}
	data := msgSegData(*optional)
	data["url"] = url
	data["title"] = title
	data["audio"] = audioUrl

	return MessageSegment{
		Type: "custom",
		Data: data,
	}
}

// 回复
func (f messageSegmentFactory) Reply(id int64) MessageSegment {
	return MessageSegment{
		Type: "reply",
		Data: msgSegData{
			"id": strconv.FormatInt(id, 10),
		},
	}
}

// 合并转发节点
func (f messageSegmentFactory) Node(id int64) MessageSegment {
	return MessageSegment{
		Type: "node",
		Data: msgSegData{
			"id": strconv.FormatInt(id, 10),
		},
	}
}

// 合并转发自定义节点
func (f messageSegmentFactory) NodeCustom(userId int64, nickname string, content Message) MessageSegment {
	return MessageSegment{
		Type: "node",
		Data: msgSegData{
			"user_id":  strconv.FormatInt(userId, 10),
			"nickname": nickname,
			"content":  content,
		},
	}
}

// XML 消息
func (f messageSegmentFactory) XML(xml string) MessageSegment {
	return MessageSegment{
		Type: "xml",
		Data: msgSegData{
			"data": xml,
		},
	}
}

// JSON 消息
func (f messageSegmentFactory) JSON(json string) MessageSegment {
	return MessageSegment{
		Type: "json",
		Data: msgSegData{
			"data": json,
		},
	}
}

//  文本转语音
func (f messageSegmentFactory) TTS(text string) MessageSegment {
	return MessageSegment{
		Type: "tts",
		Data: msgSegData{
			"text": text,
		},
	}
}
