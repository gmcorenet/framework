package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_GETRoute(t *testing.T) {
	r := New()
	_, err := r.GET("/hello", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("failed to register route: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := New()
	_, err := r.GET("/hello", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("failed to register route: %v", err)
	}

	r.SetMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	req := httptest.NewRequest(http.MethodPost, "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", w.Code)
	}
}
