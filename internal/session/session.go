package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"
)

type Session interface {
	ID() string
	Get(key string) interface{}
	Set(key string, value interface{})
	Remove(key string)
	Has(key string) bool
	Keys() []string
	Clear()
	Flush()
	Destroy()
}

type Store interface {
	New(sid string) (Session, error)
	Get(sid string) (Session, error)
	Save(session Session) error
	Delete(sid string) error
	Cleanup(maxLifetime time.Duration) error
}

type SessionStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
	maxAge   time.Duration
}

func NewSessionStore(maxAge time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]Session),
		maxAge:   maxAge,
	}
}

func (s *SessionStore) New(sid string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ss := newSession(sid)
	s.sessions[sid] = ss
	return ss, nil
}

func (s *SessionStore) Get(sid string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, ok := s.sessions[sid]; ok {
		return sess, nil
	}
	return nil, nil
}

func (s *SessionStore) Save(session Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID()] = session
	return nil
}

func (s *SessionStore) Delete(sid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sid)
	return nil
}

func (s *SessionStore) Cleanup(maxLifetime time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, sess := range s.sessions {
		if ss, ok := sess.(*session); ok {
			if now.Sub(ss.CreatedAt()) > maxLifetime {
				delete(s.sessions, id)
			}
		}
	}
	return nil
}

func (s *SessionStore) GetAll() map[string]Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]Session, len(s.sessions))
	for k, v := range s.sessions {
		result[k] = v
	}
	return result
}

func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

type session struct {
	id        string
	values    map[string]interface{}
	flashes   map[string][]string
	createdAt time.Time
	modified  bool
	mu        sync.RWMutex
}

func newSession(id string) *session {
	return &session{
		id:        id,
		values:    make(map[string]interface{}),
		flashes:   make(map[string][]string),
		createdAt: time.Now(),
	}
}

func (s *session) ID() string {
	return s.id
}

func (s *session) Get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[key]
}

func (s *session) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
	s.modified = true
}

func (s *session) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	s.modified = true
}

func (s *session) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.values[key]
	return ok
}

func (s *session) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.values))
	for k := range s.values {
		keys = append(keys, k)
	}
	return keys
}

func (s *session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]interface{})
	s.modified = true
}

func (s *session) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]interface{})
	s.flashes = make(map[string][]string)
	s.modified = true
}

func (s *session) Destroy() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]interface{})
	s.flashes = make(map[string][]string)
}

func (s *session) AddFlash(level, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flashes[level] = append(s.flashes[level], message)
	s.modified = true
}

func (s *session) GetFlashes() map[string][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	flashes := make(map[string][]string)
	for k, v := range s.flashes {
		flashes[k] = v
	}
	s.flashes = make(map[string][]string)
	s.modified = true
	return flashes
}

func (s *session) HasFlashes() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.flashes) > 0
}

func (s *session) GetFlash(level string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := s.flashes[level]
	delete(s.flashes, level)
	s.modified = true
	return messages
}

func (s *session) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

func (s *session) CreatedAt() time.Time {
	return s.createdAt
}

type Manager struct {
	store        Store
	name         string
	lifetime     time.Duration
	cookiePath   string
	cookieDomain string
	cookieSec    bool
	httpOnly     bool
	sameSite     http.SameSite
}

func NewManager(store Store, name string, lifetime time.Duration) *Manager {
	return &Manager{
		store:      store,
		name:       name,
		lifetime:   lifetime,
		cookiePath: "/",
		httpOnly:   true,
		sameSite:   http.SameSiteLaxMode,
	}
}

func (m *Manager) SetCookiePath(path string) *Manager {
	m.cookiePath = path
	return m
}

func (m *Manager) SetCookieDomain(domain string) *Manager {
	m.cookieDomain = domain
	return m
}

func (m *Manager) SetCookieSecure(secure bool) *Manager {
	m.cookieSec = secure
	return m
}

func (m *Manager) SetHttpOnly(httpOnly bool) *Manager {
	m.httpOnly = httpOnly
	return m
}

func (m *Manager) SetSameSite(sameSite http.SameSite) *Manager {
	m.sameSite = sameSite
	return m
}

func (m *Manager) Start(w http.ResponseWriter, r *http.Request) (Session, error) {
	cookie, err := r.Cookie(m.name)
	var sid string

	if err != nil {
		sid, err = m.generateSid()
		if err != nil {
			return nil, err
		}
	} else {
		sid = cookie.Value
		if sid == "" {
			sid, err = m.generateSid()
			if err != nil {
				return nil, err
			}
		}
	}

	session, err := m.store.Get(sid)
	if err != nil || session == nil {
		session, err = m.store.New(sid)
		if err != nil {
			return nil, err
		}
	}

	m.saveSession(w, session)
	return session, nil
}

func (m *Manager) saveSession(w http.ResponseWriter, session Session) {
	cookie := &http.Cookie{
		Name:     m.name,
		Value:    session.ID(),
		Path:     m.cookiePath,
		Domain:   m.cookieDomain,
		MaxAge:   int(m.lifetime.Seconds()),
		Secure:   m.cookieSec,
		HttpOnly: m.httpOnly,
		SameSite: m.sameSite,
	}
	http.SetCookie(w, cookie)
}

func (m *Manager) Save(w http.ResponseWriter, session Session) error {
	if err := m.store.Save(session); err != nil {
		return err
	}
	m.saveSession(w, session)
	return nil
}

func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.name)
	if err != nil {
		return nil
	}

	sid := cookie.Value
	if sid != "" {
		if err := m.store.Delete(sid); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.name,
		Value:    "",
		Path:     m.cookiePath,
		Domain:   m.cookieDomain,
		MaxAge:   -1,
		Secure:   m.cookieSec,
		HttpOnly: m.httpOnly,
		SameSite: m.sameSite,
	})
	return nil
}

func (m *Manager) Regenerate(w http.ResponseWriter, r *http.Request) (Session, error) {
	cookie, err := r.Cookie(m.name)
	if err != nil {
		return nil, err
	}

	session, err := m.store.Get(cookie.Value)
	if err != nil || session == nil {
		return nil, err
	}

	newSid, err := m.generateSid()
	if err != nil {
		return nil, err
	}

	m.store.Delete(session.ID())

	newSession, err := m.store.New(newSid)
	if err != nil {
		return nil, err
	}

	for _, key := range session.Keys() {
		newSession.Set(key, session.Get(key))
	}

	m.saveSession(w, newSession)
	return newSession, nil
}

func (m *Manager) generateSid() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type contextKey string

const SessionContextKey contextKey = "gmcore_session"

func SaveToContext(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

func FromContext(ctx context.Context) Session {
	if session, ok := ctx.Value(SessionContextKey).(Session); ok {
		return session
	}
	return nil
}

type Middleware struct {
	manager *Manager
}

func NewMiddleware(manager *Manager) *Middleware {
	return &Middleware{manager: manager}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.manager.Start(w, r)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		ctx := SaveToContext(r.Context(), session)

		srw := &sessionResponseWriter{
			ResponseWriter: w,
			session:       session,
			manager:       m.manager,
			written:       false,
		}

		next.ServeHTTP(srw, r.WithContext(ctx))

		if srw.written {
			m.manager.Save(srw, session)
		}
	})
}

type sessionResponseWriter struct {
	http.ResponseWriter
	session       Session
	manager       *Manager
	written       bool
}

func (w *sessionResponseWriter) Write(data []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(data)
}

func (w *sessionResponseWriter) WriteHeader(statusCode int) {
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)
}
