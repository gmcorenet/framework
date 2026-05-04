package kernel

import (
	"context"
	"net/http"
	"time"
)

const (
	EventRequest             = "kernel.request"
	EventController          = "kernel.controller"
	EventControllerFinish    = "kernel.controller.finish"
	EventResponse            = "kernel.response"
	EventException           = "kernel.exception"
	EventTerminate           = "kernel.terminate"
	EventBoot                = "kernel.boot"
)

type KernelEvent struct {
	Timestamp time.Time
	Context   context.Context
	Request   *http.Request
	Response  http.ResponseWriter
	Error     error
	Data      map[string]interface{}
}

func NewKernelEvent(ctx context.Context, req *http.Request, w http.ResponseWriter) *KernelEvent {
	return &KernelEvent{
		Timestamp: time.Now(),
		Context:   ctx,
		Request:   req,
		Response:  w,
		Data:      make(map[string]interface{}),
	}
}

type KernelEventListener interface {
	OnKernelEvent(event *KernelEvent) error
}

type EventSubscriber struct {
	Callbacks map[string]func(*KernelEvent) error
}

func NewEventSubscriber() *EventSubscriber {
	return &EventSubscriber{
		Callbacks: make(map[string]func(*KernelEvent) error),
	}
}

func (s *EventSubscriber) On(event string, callback func(*KernelEvent) error) {
	s.Callbacks[event] = callback
}

func (s *EventSubscriber) GetCallback(event string) func(*KernelEvent) error {
	if cb, ok := s.Callbacks[event]; ok {
		return cb
	}
	return nil
}

type eventManager struct {
	subscribers map[string][]*EventSubscriber
}

func newEventManager() *eventManager {
	return &eventManager{
		subscribers: make(map[string][]*EventSubscriber),
	}
}

func (em *eventManager) Subscribe(event string, subscriber *EventSubscriber) {
	em.subscribers[event] = append(em.subscribers[event], subscriber)
}

func (em *eventManager) Dispatch(ctx context.Context, event string, ke *KernelEvent) {
	ke.Context = ctx
	if subscribers, ok := em.subscribers[event]; ok {
		for _, sub := range subscribers {
			if cb := sub.GetCallback(event); cb != nil {
				if err := cb(ke); err != nil {
					// Log error but don't stop dispatch
				}
			}
		}
	}
}

func (em *eventManager) HasSubscribers(event string) bool {
	if subs, ok := em.subscribers[event]; ok && len(subs) > 0 {
		return true
	}
	return false
}
