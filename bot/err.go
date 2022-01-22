package bot

import "errors"

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrMissingAnonymous   = errors.New("missing anonymous")
)
