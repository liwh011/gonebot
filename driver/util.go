package driver

import "github.com/tidwall/gjson"

// 对还未发送就出错的请求生成一个错误响应
func makeErrorResponse(req request, err error) response {
	return response{
		Status:  "failed",
		Msg:     err.Error(),
		RetCode: -1,
		Echo:    req.Echo,
	}
}

func isResponse(msg gjson.Result) bool {
	return msg.Get("status").Exists() && msg.Get("retcode").Exists()
}
