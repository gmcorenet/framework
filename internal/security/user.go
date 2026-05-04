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
	GetRoles() []string
	IsEqual(user UserInterface) bool
	EraseCredentials()
}

type Role interface {
	GetRole() string
}

type AdvancedUserInterface interface {
	UserInterface
	IsEqual(user AdvancedUserInterface) bool
	IsAccountNonExpired() bool
	IsAccountNonLocked() bool
	IsCredentialsNonExpired() bool
	IsEnabled() bool
}
