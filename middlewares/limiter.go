package middlewares

import (
	"time"

	"github.com/liwh011/gonebot"
)

type FrequencyLimiter struct {
	Cd        int                               // 冷却时间，秒
	KeyFunc   func(ctx *gonebot.Context) string // 标识一次调用的函数
	OnFail    func(ctx *gonebot.Context)        // 如果被限制，调用该方法
	lastCalls map[string]time.Time
}

func (f *FrequencyLimiter) Handle(ctx *gonebot.Context, action *gonebot.Action) {
	if f.Cd <= 0 {
		return
	}
	// lazy init
	if f.lastCalls == nil {
		f.lastCalls = make(map[string]time.Time)
	}

	now := time.Now()
	key := f.KeyFunc(ctx)
	if last, ok := f.lastCalls[key]; ok {
		if now.Sub(last) < time.Duration(f.Cd)*time.Second {
			action.AbortHandler()
			if f.OnFail != nil {
				f.OnFail(ctx)
			}
		}
		return
	}
	f.lastCalls[key] = now
}

func NewFrequencyLimiter(cd int, keyFunc func(ctx *gonebot.Context) string) *FrequencyLimiter {
	return &FrequencyLimiter{
		Cd:      cd,
		KeyFunc: keyFunc,
	}
}

func NewSessionFreqLimiter(cd int) *FrequencyLimiter {
	return NewFrequencyLimiter(cd, func(ctx *gonebot.Context) string {
		return ctx.Event.(gonebot.I_MessageEvent).GetSessionId()
	})
}
