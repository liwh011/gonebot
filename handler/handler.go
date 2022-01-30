package handler

type HandleFunc func(*Context)

type EventHandler struct {
	handler       HandleFunc
	preconditions []func(*Context) bool
	priority      int
	block         bool
}

func (h *EventHandler) Handle(ctx *Context) {
	for _, precondition := range h.preconditions {
		if !precondition(ctx) {
			return
		}
	}
	h.handler(ctx)
}

func (h *EventHandler) SetPriority(priority int) *EventHandler {
	h.priority = priority
	return h
}

func (h *EventHandler) SetBlock(block bool) *EventHandler {
	h.block = block
	return h
}

func (h *EventHandler) SetHandler(handler HandleFunc) *EventHandler {
	h.handler = handler
	return h
}

func (h *EventHandler) AddPrecondition(precondition func(*Context) bool) *EventHandler {
	h.preconditions = append(h.preconditions, precondition)
	return h
}