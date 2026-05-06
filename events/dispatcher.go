package events

import (
	"context"
	"fmt"
	"sync"

	gmcore_events "github.com/gmcorenet/sdk/gmcore-events"
)

type Event = interface{}
type EventHandler = gmcore_events.Listener
type Listener = gmcore_events.Listener
type Unsubscribe = gmcore_events.Unsubscribe

type entry struct {
	id      int
	handler EventHandler
	unsub   Unsubscribe
}

type EventDispatcher struct {
	bus     *gmcore_events.Bus
	mu      sync.Mutex
	entries map[string][]*entry
	nextID  int
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		bus:     gmcore_events.NewBus(),
		entries: make(map[string][]*entry),
		nextID:  1,
	}
}

func (d *EventDispatcher) On(event string, handler EventHandler) func() {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := d.nextID
	d.nextID++

	unsub := d.bus.Subscribe(event, handler)
	e := &entry{id: id, handler: handler, unsub: unsub}
	d.entries[event] = append(d.entries[event], e)

	return func() {
		d.removeEntry(event, id)
	}
}

func (d *EventDispatcher) Off(event string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ptr := fmt.Sprintf("%p", handler)
	entries := d.entries[event]
	filtered := make([]*entry, 0, len(entries))
	for _, e := range entries {
		if fmt.Sprintf("%p", e.handler) != ptr {
			filtered = append(filtered, e)
		} else {
			e.unsub()
		}
	}
	d.entries[event] = filtered
}

func (d *EventDispatcher) removeEntry(event string, id int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entries := d.entries[event]
	filtered := make([]*entry, 0, len(entries))
	for _, e := range entries {
		if e.id != id {
			filtered = append(filtered, e)
		} else {
			e.unsub()
		}
	}
	d.entries[event] = filtered
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event string, payload interface{}) error {
	return d.bus.Dispatch(ctx, event, payload)
}

func (d *EventDispatcher) DispatchCollect(ctx context.Context, event string, payload interface{}) []error {
	return d.bus.DispatchCollect(ctx, event, payload)
}

func (d *EventDispatcher) HasListeners(event string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries[event]) > 0
}

func (d *EventDispatcher) ListenerCount(event string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries[event])
}

func (d *EventDispatcher) Clear(event string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.bus.UnsubscribeAll(event)
	delete(d.entries, event)
}

func (d *EventDispatcher) ClearAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for event := range d.entries {
		d.bus.UnsubscribeAll(event)
	}
	d.entries = make(map[string][]*entry)
}
