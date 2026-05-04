package events

import (
	"context"
	"sync"
	"time"
)

type EventHandler func(ctx context.Context, payload interface{}) error

type listenerEntry struct {
	id      int
	handler EventHandler
}

type Event struct {
	Name      string
	Payload   interface{}
	Context   context.Context
	Timestamp time.Time
}

type EventDispatcher struct {
	handlers map[string][]listenerEntry
	mu       sync.RWMutex
	nextID   int
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]listenerEntry),
		nextID:   1,
	}
}

func (d *EventDispatcher) On(event string, handler EventHandler) func() {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := d.nextID
	d.nextID++

	d.handlers[event] = append(d.handlers[event], listenerEntry{
		id:      id,
		handler: handler,
	})

	return func() {
		d.removeHandler(event, id)
	}
}

func (d *EventDispatcher) Off(event string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, entry := range d.handlers[event] {
		if entry.handler == handler {
			d.handlers[event] = append(d.handlers[event][:i], d.handlers[event][i+1:]...)
			break
		}
	}
}

func (d *EventDispatcher) removeHandler(event string, id int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	handlers := d.handlers[event]
	for i, entry := range handlers {
		if entry.id == id {
			d.handlers[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event string, payload interface{}) error {
	d.mu.RLock()
	entries := make([]listenerEntry, len(d.handlers[event]))
	copy(entries, d.handlers[event])
	d.mu.RUnlock()

	e := Event{
		Name:      event,
		Payload:   payload,
		Context:   ctx,
		Timestamp: time.Now(),
	}

	var lastErr error
	for _, entry := range entries {
		if err := entry.handler(ctx, e); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (d *EventDispatcher) DispatchCollect(ctx context.Context, event string, payload interface{}) []error {
	d.mu.RLock()
	entries := make([]listenerEntry, len(d.handlers[event]))
	copy(entries, d.handlers[event])
	d.mu.RUnlock()

	e := Event{
		Name:      event,
		Payload:   payload,
		Context:   ctx,
		Timestamp: time.Now(),
	}

	errors := make([]error, 0)
	for _, entry := range entries {
		if err := entry.handler(ctx, e); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (d *EventDispatcher) HasListeners(event string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[event]) > 0
}

func (d *EventDispatcher) ListenerCount(event string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[event])
}

func (d *EventDispatcher) Clear(event string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.handlers, event)
}

func (d *EventDispatcher) ClearAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = make(map[string][]listenerEntry)
}
