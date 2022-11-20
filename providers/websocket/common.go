package websocket

import "github.com/tidwall/gjson"

type request struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   uint64      `json:"echo"`
}

type response struct {
	Status  string       `json:"status"` // 'ok' or 'failed'
	Data    gjson.Result `json:"data"`
	Msg     string       `json:"msg"`     // error message
	Wording string       `json:"wording"` // error message in Chinese
	RetCode int64        `json:"retcode"` // error code, 0 for success
	Echo    uint64       `json:"echo"`    // 回复消息的序列号，用于匹配
}

func isApiResponse(msg gjson.Result) bool {
	return msg.Get("status").Exists() && msg.Get("retcode").Exists()
}
