package middlewares

import (
	"sync"
	"time"

	"github.com/liwh011/gonebot"
)

// 限制调用频率
type FrequencyLimiter struct {
	Cd        int                               // 冷却时间，秒
	KeyFunc   func(ctx *gonebot.Context) string // 标识一次调用的函数
	onFail    func(ctx *gonebot.Context)        // 如果被限制，调用该方法
	lastCalls map[string]time.Time
	mu        sync.Mutex
}

func (f *FrequencyLimiter) Handle(ctx *gonebot.Context) bool {
	if f.Cd <= 0 {
		return true
	}
	key := f.KeyFunc(ctx)

	f.mu.Lock()
	// lazy init
	if f.lastCalls == nil {
		f.lastCalls = make(map[string]time.Time)
	}

	now := time.Now()
	last, ok := f.lastCalls[key]                                     // 上次调用时间
	ready := !ok || now.Sub(last) >= time.Duration(f.Cd)*time.Second // 冷却时间是否就绪
	if ready {
		f.lastCalls[key] = now // 刷新调用时间为现在
	}
	f.mu.Unlock()

	if !ready {
		if f.onFail != nil {
			f.onFail(ctx)
		}
		return false
	}

	return true
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
	mu        sync.Mutex
}

func (d *DailyTimesLimiter) Handle(ctx *gonebot.Context) bool {
	if d.Times <= 0 {
		return true
	}

	// 超过重置时间，重置次数，将重置时间+1天
	if time.Now().After(d.resetTime) {
		d.callTimes = nil
		d.resetTime = d.resetTime.AddDate(0, 0, 1)
	}

	key := d.KeyFunc(ctx)

	d.mu.Lock()
	// lazy init
	if d.callTimes == nil {
		d.callTimes = make(map[string]int)
	}

	times, ok := d.callTimes[key]    // 当天已经调用的次数
	exceed := ok && times >= d.Times // 是否超过次数限制
	if !exceed {
		d.callTimes[key] = times + 1
	}
	d.mu.Unlock()

	if exceed {
		if d.onFail != nil {
			d.onFail(ctx)
		}
		return false
	}

	return true
}

// 设置触发限制后执行的回调方法
func (d *DailyTimesLimiter) OnFail(callback func(ctx *gonebot.Context)) *DailyTimesLimiter {
	d.onFail = callback
	return d
}

func (d *DailyTimesLimiter) GetResetTime() time.Time {
	return d.resetTime
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
