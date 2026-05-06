package kernel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKernel_NewDefaultConfig(t *testing.T) {
	k := New(nil)
	if k == nil {
		t.Fatal("expected kernel instance")
	}
	if k.Config() == nil {
		t.Fatal("expected kernel config")
	}
	if k.Config().Port == "" {
		t.Fatal("expected default port to be set")
	}
}

func TestKernel_BootstrapAndServeHTTP(t *testing.T) {
	k := New(nil)
	k.GET("/ping", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := k.Bootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}
