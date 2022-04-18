package gonebot

type Service struct {
	Handler
	name            string
	visible         bool
	doc             string
	enabled         bool
	enableOnDefault bool
	groupState      map[int64]bool
}

func newService(name string) *Service {
	sv := Service{
		name:            name,
		visible:         true,
		doc:             "",
		enabled:         true,
		enableOnDefault: true,
		groupState:      make(map[int64]bool),
	}
	sv.Use(func(c *Context) bool {
		gid, exist := getEventField(c.Event, "GroupId")
		if exist {
			if !sv.IsEnabled(gid.(int64)) {
				return false
			}
		}
		return true
	})
	return &sv
}

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
