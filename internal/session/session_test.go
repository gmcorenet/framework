package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManager_Regenerate(t *testing.T) {
	store := NewSessionStore(time.Hour)
	manager := NewManager(store, "test_session", time.Hour)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	oldSession, err := manager.Start(w, r)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	oldSid := oldSession.ID()
	oldSession.Set("key", "value")

	rWithCookie := addSessionCookie(r, w.Result(), "test_session")
	w2 := httptest.NewRecorder()

	newSession, err := manager.Regenerate(w2, rWithCookie)
	if err != nil {
		t.Fatalf("Regenerate failed: %v", err)
	}

	if newSession.ID() == oldSid {
		t.Error("new session should have different ID")
	}

	if newSession.Get("key") != "value" {
		t.Error("new session should have copied data from old session")
	}
}

func TestManager_Start(t *testing.T) {
	store := NewSessionStore(time.Hour)
	manager := NewManager(store, "test_session", time.Hour)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	session1, err := manager.Start(w, r)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	session1.Set("user", "john")

	r2 := addSessionCookie(r, w.Result(), "test_session")
	w2 := httptest.NewRecorder()

	session2, err := manager.Start(w2, r2)
	if err != nil {
		t.Fatalf("Start failed (2nd call): %v", err)
	}

	if session2.ID() != session1.ID() {
		t.Error("second Start call should return same session")
	}

	if session2.Get("user") != "john" {
		t.Error("session data should persist")
	}
}

func TestManager_Destroy(t *testing.T) {
	store := NewSessionStore(time.Hour)
	manager := NewManager(store, "test_session", time.Hour)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	session, _ := manager.Start(w, r)
	session.Set("user", "john")

	rWithCookie := addSessionCookie(r, w.Result(), "test_session")
	w2 := httptest.NewRecorder()

	err := manager.Destroy(w2, rWithCookie)
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	if store.Count() != 0 {
		t.Error("session store should be empty after Destroy")
	}
}

func TestSessionStore_Operations(t *testing.T) {
	store := NewSessionStore(time.Hour)

	session, err := store.New("test-sid")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	session.Set("key1", "value1")
	session.Set("key2", 42)

	if session.Get("key1") != "value1" {
		t.Error("key1 should be value1")
	}

	if session.Get("key2") != 42 {
		t.Error("key2 should be 42")
	}

	if !session.Has("key1") {
		t.Error("Has(key1) should be true")
	}

	keys := session.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() returned %d keys, want 2", len(keys))
	}

	session.Remove("key1")
	if session.Has("key1") {
		t.Error("key1 should be removed")
	}

	session.Clear()
	if session.Has("key2") {
		t.Error("key2 should be cleared")
	}

	session.Set("data", "exists")
	session.Flush()
	if session.Has("data") {
		t.Error("data should be flushed")
	}
}

func addSessionCookie(r *http.Request, response *http.Response, name string) *http.Request {
	for _, c := range response.Cookies() {
		if c.Name == name {
			r.AddCookie(c)
			break
		}
	}
	return r
}