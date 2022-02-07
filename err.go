package gonebot

import "errors"

var (
	ErrInvalidMessageType = errors.New("不正确的message type")
	ErrMissingAnonymous   = errors.New("缺少anonymous对象")
)
