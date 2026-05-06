package security

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticator_AuthenticateSuccess(t *testing.T) {
	provider := NewUserProvider()
	hasher := NewPlainPasswordHasher()
	provider.AddUser(NewUser("alice", "secret", []string{"ROLE_USER"}))

	auth := NewAuthenticator(provider, hasher)
	req := httptest.NewRequest("POST", "/login", strings.NewReader("_username=alice&_password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	token, err := auth.Authenticate(req)
	if err != nil {
		t.Fatalf("expected successful auth, got error: %v", err)
	}
	if token == nil || token.User == nil {
		t.Fatal("expected token with user")
	}
}

func TestAuthenticator_AuthenticateFailure(t *testing.T) {
	provider := NewUserProvider()
	hasher := NewPlainPasswordHasher()
	provider.AddUser(NewUser("alice", "secret", []string{"ROLE_USER"}))

	auth := NewAuthenticator(provider, hasher)
	req := httptest.NewRequest("POST", "/login", strings.NewReader("_username=alice&_password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err := auth.Authenticate(req)
	if err == nil {
		t.Fatal("expected authentication error")
	}
}
