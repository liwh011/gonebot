package gonebot

import (
	"regexp"

	"github.com/liwh011/gonebot/event"
	"github.com/liwh011/gonebot/handler"
)

type Service struct {
	name            string
	visible         bool
	doc             string
	enabled         bool
	enableOnDefault bool
	groupState      map[int64]bool
}

func NewService(name string) *Service {
	sv := Service{
		name:            name,
		visible:         true,
		doc:             "",
		enabled:         true,
		enableOnDefault: true,
		groupState:      make(map[int64]bool),
	}
	return &sv
}

// type ServiceMiddleware interface {
// }

func (sv *Service) Enable() {
	sv.enabled = true
}

func (sv *Service) Disable() {
	sv.enabled = false
}

func (sv *Service) EnableInGroup(groupId int64) {
	sv.groupState[groupId] = true
}

func (sv *Service) DisableInGroup(groupId int64) {
	sv.groupState[groupId] = false
}

func (sv *Service) IsEnabled(groupId int64) bool {
	if !sv.enabled {
		return false
	}
	if group, ok := sv.groupState[groupId]; ok {
		return group
	}

	return sv.enableOnDefault
}

func (sv *Service) On(eventName event.EventName, cond ...handler.Condition) *handler.EventHandler {
	h := handler.EventHandler{}
	h.AddPrecondition(func(c *handler.Context) bool {
		v, exist := event.GetEventField(c.Event, "GroupId")
		if exist {
			return sv.IsEnabled(v.(int64))
		}

		return true
	})
	h.AddPrecondition(handler.EventType(eventName))
	h.AddPrecondition(cond...)
	h.SetPriority(-1)

	handlers = append(handlers, &h)
	return &h
}

func (sv *Service) OnMessage(cond ...handler.Condition) *handler.EventHandler {
	return sv.On(event.EVENT_NAME_MESSAGE, cond...)
}

func (sv *Service) OnStartsWith(prefix ...string) *handler.EventHandler {
	return sv.OnMessage(handler.StartsWith(prefix...))
}

func (sv *Service) OnEndsWith(suffix ...string) *handler.EventHandler {
	return sv.OnMessage(handler.EndsWith(suffix...))
}

func (sv *Service) OnCommand(command ...string) *handler.EventHandler {
	return sv.OnMessage(handler.Command("", command...)).SetBlock(true).SetPriority(1)
}

func (sv *Service) OnRegex(regex regexp.Regexp) *handler.EventHandler {
	return sv.OnMessage(handler.Regex(regex))
}
