package events

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestEventDispatcher_DispatchAndUnsubscribe(t *testing.T) {
	d := NewEventDispatcher()
	var called int32

	unsub := d.On("app.boot", func(ctx context.Context, payload interface{}) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	if err := d.Dispatch(context.Background(), "app.boot", "ok"); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}

	unsub()
	_ = d.Dispatch(context.Background(), "app.boot", "ok")
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected no extra calls after unsubscribe, got %d", called)
	}
}

func TestEventDispatcher_ListenerIntrospection(t *testing.T) {
	d := NewEventDispatcher()
	if d.HasListeners("x") {
		t.Fatal("expected no listeners")
	}

	d.On("x", func(ctx context.Context, payload interface{}) error { return nil })
	if !d.HasListeners("x") {
		t.Fatal("expected listeners")
	}
	if d.ListenerCount("x") != 1 {
		t.Fatalf("expected listener count 1, got %d", d.ListenerCount("x"))
	}

	d.Clear("x")
	if d.HasListeners("x") {
		t.Fatal("expected listeners to be cleared")
	}
}
