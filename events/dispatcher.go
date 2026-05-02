package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type EventHandler func(ctx context.Context, payload interface{}) error

type Event struct {
	Name      string
	Payload   interface{}
	Context   context.Context
	Timestamp time.Time
}

type EventDispatcher struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (d *EventDispatcher) On(event string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[event] = append(d.handlers[event], handler)
}

func (d *EventDispatcher) Off(event string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	handlers := d.handlers[event]
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			d.handlers[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event string, payload interface{}) error {
	d.mu.RLock()
	handlers := make([]EventHandler, len(d.handlers[event]))
	copy(handlers, d.handlers[event])
	d.mu.RUnlock()

	e := Event{
		Name:      event,
		Payload:   payload,
		Context:   ctx,
		Timestamp: time.Now(),
	}

	for _, handler := range handlers {
		if err := handler(ctx, e); err != nil {
			return err
		}
	}

	return nil
}

func (d *EventDispatcher) HasListeners(event string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[event]) > 0
}

