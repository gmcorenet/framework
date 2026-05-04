package security

import (
	"context"
)

type VoterInterface interface {
	Vote(user UserInterface, attribute string, subject interface{}) int
}

const (
	ACCESS_GRANTED = 1
	ACCESS_ABSTAIN = 0
	ACCESS_DENIED  = -1
)

type Voter struct {
	attributes []string
}

func NewVoter() *Voter {
	return &Voter{}
}

func (v *Voter) GetAttributes() []string {
	return v.attributes
}

func (v *Voter) SetAttributes(attributes []string) {
	v.attributes = attributes
}

func (v *Voter) Vote(user UserInterface, attribute string, subject interface{}) int {
	if !v.supportsAttribute(attribute) {
		return ACCESS_ABSTAIN
	}

	if user == nil {
		return ACCESS_DENIED
	}

	return v.doVote(user, attribute, subject)
}

func (v *Voter) supportsAttribute(attribute string) bool {
	if len(v.attributes) == 0 {
		return true
	}

	for _, attr := range v.attributes {
		if attr == attribute {
			return true
		}
	}
	return false
}

func (v *Voter) doVote(user UserInterface, attribute string, subject interface{}) int {
	return ACCESS_ABSTAIN
}

type RoleVoter struct {
	*Voter
	rolePrefix string
}

func NewRoleVoter() *RoleVoter {
	v := &RoleVoter{
		Voter:      NewVoter(),
		rolePrefix: "ROLE_",
	}
	v.SetAttributes([]string{"ROLE_USER", "ROLE_ADMIN", "ROLE_SUPER_ADMIN"})
	return v
}

func (v *RoleVoter) doVote(user UserInterface, attribute string, subject interface{}) int {
	roles := user.GetRoles()

	for _, role := range roles {
		if role == v.rolePrefix+attribute || role == attribute {
			return ACCESS_GRANTED
		}
	}

	return ACCESS_DENIED
}

type AuthenticatedVoter struct {
	*Voter
}

func NewAuthenticatedVoter() *AuthenticatedVoter {
	return &AuthenticatedVoter{
		Voter: NewVoter(),
	}
}

func (v *AuthenticatedVoter) doVote(user UserInterface, attribute string, subject interface{}) int {
	roles := user.GetRoles()

	if len(roles) == 0 {
		return ACCESS_DENIED
	}

	if attribute == "IS_AUTHENTICATED_FULLY" {
		for _, role := range roles {
			if role != "IS_AUTHENTICATED_ANONYMOUSLY" {
				return ACCESS_GRANTED
			}
		}
		return ACCESS_DENIED
	}

	if attribute == "IS_AUTHENTICATED_REMEMBERED" {
		for _, role := range roles {
			if role == "IS_AUTHENTICATED_FULLY" || role == "IS_AUTHENTICATED_REMEMBERED" {
				return ACCESS_GRANTED
			}
		}
	}

	if attribute == "IS_AUTHENTICATED_ANONYMOUSLY" {
		return ACCESS_GRANTED
	}

	return ACCESS_ABSTAIN
}

type SecurityChecker struct {
	voters []VoterInterface
}

func NewSecurityChecker(voters []VoterInterface) *SecurityChecker {
	return &SecurityChecker{voters: voters}
}

func (sc *SecurityChecker) AddVoter(voter VoterInterface) {
	sc.voters = append(sc.voters, voter)
}

func (sc *SecurityChecker) IsGranted(ctx context.Context, attribute string, subject interface{}) bool {
	user := UserFromContext(ctx)
	return sc.IsGrantedUser(user, attribute, subject)
}

func (sc *SecurityChecker) IsGrantedUser(user UserInterface, attribute string, subject interface{}) bool {
	for _, voter := range sc.voters {
		result := voter.Vote(user, attribute, subject)
		if result == ACCESS_DENIED {
			return false
		}
		if result == ACCESS_GRANTED {
			return true
		}
	}
	return false
}

func (sc *SecurityChecker) GetVoters() []VoterInterface {
	return sc.voters
}

var ErrInvalidCredentials = &SecurityError{code: "INVALID_CREDENTIALS", message: "Invalid credentials"}
var ErrUserNotFound = &SecurityError{code: "USER_NOT_FOUND", message: "User not found"}

type SecurityError struct {
	code    string
	message string
}

func (e *SecurityError) Error() string {
	return e.message
}

func (e *SecurityError) Code() string {
	return e.code
}
