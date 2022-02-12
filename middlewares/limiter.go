package middlewares

import (
	"time"

	"github.com/liwh011/gonebot"
)

// 限制调用频率
type FrequencyLimiter struct {
	Cd        int                               // 冷却时间，秒
	KeyFunc   func(ctx *gonebot.Context) string // 标识一次调用的函数
	onFail    func(ctx *gonebot.Context)        // 如果被限制，调用该方法
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
			if f.onFail != nil {
				f.onFail(ctx)
			}
		}
		return
	}
	f.lastCalls[key] = now
}

// 设置触发限制后执行的回调方法
func (f *FrequencyLimiter) OnFail(callback func(ctx *gonebot.Context)) *FrequencyLimiter {
	f.onFail = callback
	return f
}

func NewFrequencyLimiter(cd int, keyFunc func(ctx *gonebot.Context) string) *FrequencyLimiter {
	return &FrequencyLimiter{
		Cd:      cd,
		KeyFunc: keyFunc,
	}
}

// 以Session为粒度限制调用频率
func NewSessionFreqLimiter(cd int) *FrequencyLimiter {
	return NewFrequencyLimiter(cd, func(ctx *gonebot.Context) string {
		return ctx.Event.GetSessionId()
	})
}

// 每日次数限制
type DailyTimesLimiter struct {
	Times     int                               // 每天允许的调用次数
	KeyFunc   func(ctx *gonebot.Context) string // 标识一次调用的函数
	onFail    func(ctx *gonebot.Context)        // 如果被限制，调用该方法
	resetTime time.Time                         // 每天重置的时间
	callTimes map[string]int                    // 当天已经调用的次数
}

func (d *DailyTimesLimiter) Handle(ctx *gonebot.Context, action *gonebot.Action) {
	if d.Times <= 0 {
		return
	}

	// 超过重置时间，重置次数，将重置时间+1天
	if time.Now().After(d.resetTime) {
		d.callTimes = nil
		d.resetTime = d.resetTime.AddDate(0, 0, 1)
	}

	// lazy init
	if d.callTimes == nil {
		d.callTimes = make(map[string]int)
	}

	key := d.KeyFunc(ctx)
	if times, ok := d.callTimes[key]; ok {
		if times >= d.Times {
			action.AbortHandler()
			if d.onFail != nil {
				d.onFail(ctx)
			}
		} else {
			d.callTimes[key] = times + 1
		}
	} else {
		d.callTimes[key] = 1
	}
}

// 设置触发限制后执行的回调方法
func (d *DailyTimesLimiter) OnFail(callback func(ctx *gonebot.Context)) *DailyTimesLimiter {
	d.onFail = callback
	return d
}

// 设置重置时间（时、分）
func (d *DailyTimesLimiter) SetResetTime(h int, m int) *DailyTimesLimiter {
	// 如果设置的时间比当前时间小，则设置为下一天的这个时候；反之设置为当天
	now := time.Now()
	if now.Hour() < h || (now.Hour() == h && now.Minute() < m) {
		d.resetTime = time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.Local)
	} else {
		d.resetTime = time.Date(now.Year(), now.Month(), now.Day()+1, h, m, 0, 0, time.Local)
	}
	return d
}

// 每日次数限制
func NewDailyTimesLimiter(times int, keyFunc func(ctx *gonebot.Context) string) *DailyTimesLimiter {
	l := &DailyTimesLimiter{
		Times:   times,
		KeyFunc: keyFunc,
	}
	l.SetResetTime(4, 0)
	return l
}

// 以Session为粒度限制调用每日调用次数
func NewSessionDailyTimesLimiter(times int) *DailyTimesLimiter {
	return NewDailyTimesLimiter(times, func(ctx *gonebot.Context) string {
		return ctx.Event.GetSessionId()
	})
}
