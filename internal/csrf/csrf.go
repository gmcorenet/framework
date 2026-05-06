package csrf

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gmcorenet/framework/container"
	"github.com/gmcorenet/framework/router"
	"github.com/gmcorenet/framework/routing"
)

func init() {
	routing.RegisterMiddlewareProvider(func(ctr *container.Container, r *router.Router) (func(http.Handler) http.Handler, bool) {
		secret := "change-me"
		if s, err := ctr.Get("csrf_secret"); err == nil {
			if str, ok := s.(string); ok && str != "" {
				secret = str
			}
		}
		tm := NewTokenManager(secret)
		mw := NewMiddleware(tm)
		return mw.Handler, true
	})
}

var (
	ErrNoToken      = errors.New("csrf: token not found")
	ErrInvalidToken  = errors.New("csrf: invalid token")
	ErrNoSession     = errors.New("csrf: session not available")
)

type TokenManager struct {
	secret       []byte
	sessionName  string
	cookieName   string
	tokenLength  int
	tokenCache   map[string]*tokenEntry
	cacheMu      sync.RWMutex
	lifetime     time.Duration
	cookiePath   string
	cookieDomain string
	secure       bool
	httpOnly     bool
	sameSite     http.SameSite
	stopCleanup  chan struct{}
	maxCacheSize int
}

type tokenEntry struct {
	token     string
	createdAt time.Time
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret:       []byte(secret),
		sessionName:  "_csrf_token",
		cookieName:   "_csrf_cache",
		tokenLength:  32,
		tokenCache:   make(map[string]*tokenEntry),
		lifetime:    time.Hour,
		cookiePath:  "/",
		httpOnly:    true,
		sameSite:    http.SameSiteStrictMode,
		maxCacheSize: 10000,
	}
}

func (tm *TokenManager) SetSessionName(name string) *TokenManager {
	tm.sessionName = name
	return tm
}

func (tm *TokenManager) SetCookieName(name string) *TokenManager {
	tm.cookieName = name
	return tm
}

func (tm *TokenManager) SetCookiePath(path string) *TokenManager {
	tm.cookiePath = path
	return tm
}

func (tm *TokenManager) SetCookieDomain(domain string) *TokenManager {
	tm.cookieDomain = domain
	return tm
}

func (tm *TokenManager) SetSecure(secure bool) *TokenManager {
	tm.secure = secure
	return tm
}

func (tm *TokenManager) SetHttpOnly(httpOnly bool) *TokenManager {
	tm.httpOnly = httpOnly
	return tm
}

func (tm *TokenManager) SetSameSite(sameSite http.SameSite) *TokenManager {
	tm.sameSite = sameSite
	return tm
}

func (tm *TokenManager) SetLifetime(lifetime time.Duration) *TokenManager {
	tm.lifetime = lifetime
	return tm
}

func (tm *TokenManager) GenerateToken() (string, error) {
	tokenBytes := make([]byte, tm.tokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	hm := hmac.New(sha256.New, tm.secret)
	hm.Write([]byte(token))
	signature := base64.URLEncoding.EncodeToString(hm.Sum(nil))

	fullToken := token + "." + signature

	tm.cacheMu.Lock()
	if len(tm.tokenCache) >= tm.maxCacheSize {
		tm.cleanupExpiredTokensLocked()
		if len(tm.tokenCache) >= tm.maxCacheSize {
			tm.cacheMu.Unlock()
			return "", errors.New("csrf: token cache full")
		}
	}
	tm.tokenCache[fullToken] = &tokenEntry{
		token:     fullToken,
		createdAt: time.Now(),
	}
	tm.cacheMu.Unlock()

	return fullToken, nil
}

func (tm *TokenManager) ValidateToken(token string) error {
	if token == "" {
		return ErrNoToken
	}

	parts := splitToken(token)
	if len(parts) != 2 {
		return ErrInvalidToken
	}

	tokenPart := parts[0]
	signaturePart := parts[1]

	hm := hmac.New(sha256.New, tm.secret)
	hm.Write([]byte(tokenPart))
	expectedSignature := base64.URLEncoding.EncodeToString(hm.Sum(nil))

	if !hmac.Equal([]byte(signaturePart), []byte(expectedSignature)) {
		return ErrInvalidToken
	}

	tm.cacheMu.Lock()
	entry, ok := tm.tokenCache[token]
	if !ok {
		tm.cacheMu.Unlock()
		return ErrInvalidToken
	}

	if time.Since(entry.createdAt) > tm.lifetime {
		delete(tm.tokenCache, token)
		tm.cacheMu.Unlock()
		return ErrInvalidToken
	}
	delete(tm.tokenCache, token)
	tm.cacheMu.Unlock()

	return nil
}

func splitToken(token string) []string {
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return nil
}

func (tm *TokenManager) GetTokenFromRequest(r *http.Request) string {
	token := r.FormValue("_csrf_token")
	if token != "" {
		return token
	}

	token = r.Header.Get("X-CSRF-Token")
	if token != "" {
		return token
	}

	token = r.Header.Get("X-XSRF-Token")
	return token
}

func (tm *TokenManager) ValidateRequest(r *http.Request) error {
	token := tm.GetTokenFromRequest(r)
	if token == "" {
		return ErrNoToken
	}
	return tm.ValidateToken(token)
}

func (tm *TokenManager) TokenFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(tm.sessionName)
	if err != nil {
		return "", ErrNoSession
	}
	return cookie.Value, nil
}

func (tm *TokenManager) SaveTokenToSession(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := tm.GenerateToken()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     tm.sessionName,
		Value:    token,
		Path:     tm.cookiePath,
		Domain:   tm.cookieDomain,
		MaxAge:   int(tm.lifetime.Seconds()),
		Secure:   tm.secure,
		HttpOnly: tm.httpOnly,
		SameSite: tm.sameSite,
	})

	return token, nil
}

func (tm *TokenManager) ClearToken(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     tm.sessionName,
		Value:    "",
		Path:     tm.cookiePath,
		Domain:   tm.cookieDomain,
		MaxAge:   -1,
		Secure:   tm.secure,
		HttpOnly: tm.httpOnly,
		SameSite: tm.sameSite,
	})
}

func (tm *TokenManager) CleanupExpiredTokens() {
	tm.cacheMu.Lock()
	defer tm.cacheMu.Unlock()
	tm.cleanupExpiredTokensLocked()
}

func (tm *TokenManager) cleanupExpiredTokensLocked() {
	now := time.Now()
	for token, entry := range tm.tokenCache {
		if now.Sub(entry.createdAt) > tm.lifetime {
			delete(tm.tokenCache, token)
		}
	}
}

func (tm *TokenManager) StartCleanup(ctx context.Context, interval time.Duration) {
	tm.cacheMu.Lock()
	if tm.stopCleanup != nil {
		tm.cacheMu.Unlock()
		return
	}
	tm.stopCleanup = make(chan struct{})
	tm.cacheMu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tm.CleanupExpiredTokens()
			case <-tm.stopCleanup:
				return
			case <-ctx.Done():
				tm.StopCleanup()
				return
			}
		}
	}()
}

func (tm *TokenManager) StopCleanup() {
	tm.cacheMu.Lock()
	defer tm.cacheMu.Unlock()
	if tm.stopCleanup != nil {
		close(tm.stopCleanup)
		tm.stopCleanup = nil
	}
}

type contextKey string

const TokenManagerKey contextKey = "csrf_token_manager"

func SaveToContext(r *http.Request, tm *TokenManager) *http.Request {
	return r.WithContext(WithContext(r.Context(), tm))
}

func WithContext(ctx context.Context, tm *TokenManager) context.Context {
	return context.WithValue(ctx, TokenManagerKey, tm)
}

func FromContext(ctx context.Context) *TokenManager {
	if tm, ok := ctx.Value(TokenManagerKey).(*TokenManager); ok {
		return tm
	}
	return nil
}

type Middleware struct {
	tokenManager *TokenManager
	excludedPaths []string
}

func NewMiddleware(tm *TokenManager) *Middleware {
	return &Middleware{
		tokenManager: tm,
		excludedPaths: []string{},
	}
}

func (m *Middleware) AddExcludedPath(path string) {
	m.excludedPaths = append(m.excludedPaths, path)
}

func (m *Middleware) ExcludedPaths(paths []string) {
	m.excludedPaths = paths
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, path := range m.excludedPaths {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			existing, _ := m.tokenManager.TokenFromSession(r)
			if existing == "" {
				_, err := m.tokenManager.SaveTokenToSession(w, r)
				if err != nil {
					http.Error(w, "CSRF error", http.StatusInternalServerError)
					return
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		if err := m.tokenManager.ValidateRequest(r); err != nil {
			http.Error(w, "CSRF validation failed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
