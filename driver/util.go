package driver

// 对还未发送就出错的请求生成一个错误响应
func makeErrorResponse(req request, err error) message {
	return message{
		Status:  "failed",
		Msg:     err.Error(),
		RetCode: -1,
		Echo:    req.Echo,
	}
}

// 判断消息是否是回复
func isReplyMessage(msg message) bool {
	return msg.Echo.seq != 0
}
