package security

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type AuthToken struct {
	User     UserInterface
	Credentials interface{}
}

type AuthenticatorInterface interface {
	Authenticate(r *http.Request) (*AuthToken, error)
	OnAuthSuccess(w http.ResponseWriter, r *http.Request, token *AuthToken)
	OnAuthFailure(w http.ResponseWriter, r *http.Request, err error)
}

type Authenticator struct {
	userProvider UserProviderInterface
	hasher       PasswordHasherInterface
}

func NewAuthenticator(provider UserProviderInterface, hasher PasswordHasherInterface) *Authenticator {
	return &Authenticator{
		userProvider: provider,
		hasher:       hasher,
	}
}

func (a *Authenticator) Authenticate(r *http.Request) (*AuthToken, error) {
	username := r.FormValue("_username")
	password := r.FormValue("_password")

	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := a.userProvider.LoadUserByIdentifier(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !a.hasher.Verify(user.GetPassword(), password) {
		return nil, ErrInvalidCredentials
	}

	return &AuthToken{User: user}, nil
}

func (a *Authenticator) OnAuthSuccess(w http.ResponseWriter, r *http.Request, token *AuthToken) {
	if sc := SecurityContextFromContext(r.Context()); sc != nil {
		sc.SetToken(token)
	}
}

func (a *Authenticator) OnAuthFailure(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Authentication failed", http.StatusUnauthorized)
}

type UserProviderInterface interface {
	LoadUserByIdentifier(identifier string) (UserInterface, error)
}

type UserProvider struct {
	users map[string]UserInterface
}

func NewUserProvider() *UserProvider {
	return &UserProvider{users: make(map[string]UserInterface)}
}

func (p *UserProvider) AddUser(user UserInterface) {
	identifier := user.GetIdentifier()
	key := toString(identifier)
	p.users[key] = user
}

func (p *UserProvider) LoadUserByIdentifier(identifier string) (UserInterface, error) {
	if user, ok := p.users[identifier]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (p *UserProvider) GetUsers() map[string]UserInterface {
	result := make(map[string]UserInterface, len(p.users))
	for k, v := range p.users {
		result[k] = v
	}
	return result
}

type User struct {
	identifier interface{}
	roles      []string
	password   string
}

func NewUser(identifier interface{}, password string, roles []string) *User {
	return &User{
		identifier: identifier,
		password:   password,
		roles:      roles,
	}
}

func (u *User) GetIdentifier() interface{} {
	return u.identifier
}

func (u *User) GetRoles() []string {
	return u.roles
}

func (u *User) GetPassword() string {
	return u.password
}

func (u *User) EraseCredentials() {
	u.password = ""
}

func (u *User) IsEqual(user UserInterface) bool {
	return u.GetIdentifier() == user.GetIdentifier()
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprint(val)
	}
}

type ContextKeys string

const (
	SecurityUser ContextKeys = "security_user"
)

func SaveUserToContext(ctx context.Context, user UserInterface) context.Context {
	return context.WithValue(ctx, SecurityUser, user)
}

func UserFromContext(ctx context.Context) UserInterface {
	if user, ok := ctx.Value(SecurityUser).(UserInterface); ok {
		return user
	}
	return nil
}
