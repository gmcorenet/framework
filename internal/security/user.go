package security

type User interface {
	GetRoles() []string
	GetPassword() string
	EraseCredentials()
}

type RoleHierarchy interface {
	GetRoles() []string
}

type UserInterface interface {
	GetIdentifier() interface{}
	GetPassword() string
	GetRoles() []string
	IsEqual(user UserInterface) bool
	EraseCredentials()
}

type Role interface {
	GetRole() string
}

type AdvancedUserInterface interface {
	UserInterface
	IsAccountNonExpired() bool
	IsAccountNonLocked() bool
	IsCredentialsNonExpired() bool
	IsEnabled() bool
}
