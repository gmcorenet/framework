package security

import (
	"context"
	"errors"
	"net/http"
	"regexp"
)

type Firewall struct {
	name          string
	pattern       *regexp.Regexp
	entryPoint    string
	authenticators []AuthenticatorInterface
	accessMap     *AccessMap
	contextMgr    *ContextManager
	sessionMgr    *ContextManager
	voters        []Voter
}

func NewFirewall(name string, pattern string) (*Firewall, error) {
	var re *regexp.Regexp
	var err error
	if pattern != "" {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}
	return &Firewall{
		name:          name,
		pattern:       re,
		authenticators: make([]AuthenticatorInterface, 0),
		accessMap:     NewAccessMap(),
		contextMgr:    NewContextManager(),
		sessionMgr:    NewContextManager(),
		voters:        make([]Voter, 0),
	}, nil
}

func (f *Firewall) SetEntryPoint(entryPoint string) {
	f.entryPoint = entryPoint
}

func (f *Firewall) AddAuthenticator(auth AuthenticatorInterface) {
	f.authenticators = append(f.authenticators, auth)
}

func (f *Firewall) SetAccessMap(accessMap *AccessMap) {
	f.accessMap = accessMap
}

func (f *Firewall) AddVoter(voter Voter) {
	f.voters = append(f.voters, voter)
}

func (f *Firewall) SetContextManager(ctxMgr *ContextManager) {
	f.contextMgr = ctxMgr
}

func (f *Firewall) SetSessionManager(sessMgr *ContextManager) {
	f.sessionMgr = sessMgr
}

func (f *Firewall) Name() string {
	return f.name
}

func (f *Firewall) Pattern() *regexp.Regexp {
	return f.pattern
}

func (f *Firewall) EntryPoint() string {
	return f.entryPoint
}

func (f *Firewall) Authenticators() []AuthenticatorInterface {
	result := make([]AuthenticatorInterface, len(f.authenticators))
	copy(result, f.authenticators)
	return result
}

func (f *Firewall) Handle(w http.ResponseWriter, r *http.Request) (*AuthToken, error) {
	if f.pattern != nil && !f.pattern.MatchString(r.URL.Path) {
		return nil, nil
	}

	if f.entryPoint != "" && len(f.authenticators) == 0 {
		return f.entryPointAuth(w, r)
	}

	for _, authenticator := range f.authenticators {
		token, err := authenticator.Authenticate(r)
		if err == nil && token != nil {
			authenticator.OnAuthSuccess(w, r, token)

			if err := f.checkAccess(w, r, token); err != nil {
				return nil, err
			}

			return token, nil
		}
	}

	return nil, nil
}

func (f *Firewall) checkAccess(w http.ResponseWriter, r *http.Request, token *AuthToken) error {
	if token == nil || token.User == nil {
		return nil
	}

	attrs := f.accessMap.Get(r.URL.Path)
	if attrs == nil {
		return nil
	}

	for _, attr := range attrs {
		granted := false
		for _, voter := range f.voters {
			result := voter.Vote(token.User, attr, r)
			if result == ACCESS_GRANTED {
				granted = true
				break
			}
			if result == ACCESS_DENIED {
				http.Error(w, "Access Denied", http.StatusForbidden)
				return errors.New("access denied")
			}
		}
		if !granted {
			http.Error(w, "Access Denied", http.StatusForbidden)
			return errors.New("access denied")
		}
	}

	return nil
}

func (f *Firewall) entryPointAuth(w http.ResponseWriter, r *http.Request) (*AuthToken, error) {
	return nil, ErrUserNotFound
}

type ContextManager struct {
}

func NewContextManager() *ContextManager {
	return &ContextManager{}
}

type AccessMap struct {
	accessRules map[string][]string
}

func NewAccessMap() *AccessMap {
	return &AccessMap{
		accessRules: make(map[string][]string),
	}
}

func (m *AccessMap) Add(path string, attributes []string) {
	m.accessRules[path] = attributes
}

func (m *AccessMap) Get(path string) []string {
	if attrs, ok := m.accessRules[path]; ok {
		return attrs
	}
	return nil
}

func (m *AccessMap) Remove(path string) {
	delete(m.accessRules, path)
}

type FirewallManager struct {
	firewalls map[string]*Firewall
}

func NewFirewallManager() *FirewallManager {
	return &FirewallManager{
		firewalls: make(map[string]*Firewall),
	}
}

func (m *FirewallManager) Add(firewall *Firewall) {
	m.firewalls[firewall.Name()] = firewall
}

func (m *FirewallManager) Get(name string) *Firewall {
	if fw, ok := m.firewalls[name]; ok {
		return fw
	}
	return nil
}

func (m *FirewallManager) All() map[string]*Firewall {
	result := make(map[string]*Firewall, len(m.firewalls))
	for k, v := range m.firewalls {
		result[k] = v
	}
	return result
}

func (m *FirewallManager) Remove(name string) {
	delete(m.firewalls, name)
}

type SecurityContext struct {
	context.Context
	token     *AuthToken
	trusted   bool
	listeners []func(*AuthToken)
}

func NewSecurityContext(ctx context.Context) *SecurityContext {
	return &SecurityContext{
		Context: ctx,
	}
}

func (sc *SecurityContext) SetToken(token *AuthToken) {
	sc.token = token
	for _, listener := range sc.listeners {
		listener(token)
	}
}

func (sc *SecurityContext) GetToken() *AuthToken {
	return sc.token
}

func (sc *SecurityContext) IsAuthenticated() bool {
	return sc.token != nil && sc.token.User != nil
}

func (sc *SecurityContext) GetUser() UserInterface {
	if sc.token == nil {
		return nil
	}
	return sc.token.User
}

func (sc *SecurityContext) AddListener(f func(*AuthToken)) {
	sc.listeners = append(sc.listeners, f)
}

type securityContextKeyType string

const securityContextKey securityContextKeyType = "gmcore_security_context"

func SecurityContextFromContext(ctx context.Context) *SecurityContext {
	if sc, ok := ctx.Value(securityContextKey).(*SecurityContext); ok {
		return sc
	}
	return nil
}

func SaveSecurityContextToContext(ctx context.Context, sc *SecurityContext) context.Context {
	return context.WithValue(ctx, securityContextKey, sc)
}
