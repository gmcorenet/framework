package security

import (
	"crypto/subtle"
	"fmt"
)

type PasswordHasherInterface interface {
	Hash(password string) (string, error)
	Verify(hashedPassword, plainPassword string) bool
	NeedsRehash(hashedPassword string) bool
}

type PasswordHasher struct {
	options map[string]interface{}
}

func NewPasswordHasher(options map[string]interface{}) *PasswordHasher {
	if options == nil {
		options = make(map[string]interface{})
	}
	return &PasswordHasher{options: options}
}

func (p *PasswordHasher) Hash(password string) (string, error) {
	return HashPassword(password)
}

func (p *PasswordHasher) Verify(hashedPassword, plainPassword string) bool {
	return VerifyPassword(hashedPassword, plainPassword)
}

func (p *PasswordHasher) NeedsRehash(hashedPassword string) bool {
	prefix := "$2a$"
	if p.options["algorithm"] == "argon2" {
		prefix = "$argon2"
	}
	return len(hashedPassword) < 60 || !hasPrefix(hashedPassword, prefix)
}

type PlainPasswordHasher struct{}

func NewPlainPasswordHasher() *PlainPasswordHasher {
	return &PlainPasswordHasher{}
}

func (p *PlainPasswordHasher) Hash(password string) (string, error) {
	return password, nil
}

func (p *PlainPasswordHasher) Verify(hashedPassword, plainPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(plainPassword)) == 1
}

func (p *PlainPasswordHasher) NeedsRehash(hashedPassword string) bool {
	return false
}

type MigratingPasswordHasher struct {
	hashers map[string]PasswordHasherInterface
	migratedHashers map[string]PasswordHasherInterface
}

func NewMigratingPasswordHasher(hashers map[string]PasswordHasherInterface) *MigratingPasswordHasher {
	return &MigratingPasswordHasher{
		hashers: hashers,
	}
}

func (m *MigratingPasswordHasher) AddHashers(hashers map[string]PasswordHasherInterface) {
	m.migratedHashers = hashers
}

func (m *MigratingPasswordHasher) Hash(password string) (string, error) {
	if len(m.hashers) == 0 {
		return "", fmt.Errorf("no hashers available")
	}
	var hasher PasswordHasherInterface
	for _, h := range m.hashers {
		hasher = h
		break
	}
	return hasher.Hash(password)
}

func (m *MigratingPasswordHasher) Verify(hashedPassword, plainPassword string) bool {
	for _, hasher := range m.hashers {
		if hasher.Verify(hashedPassword, plainPassword) {
			return true
		}
	}
	for _, hasher := range m.migratedHashers {
		if hasher.Verify(hashedPassword, plainPassword) {
			return true
		}
	}
	return false
}

func (m *MigratingPasswordHasher) NeedsRehash(hashedPassword string) bool {
	for _, hasher := range m.hashers {
		if hasher.NeedsRehash(hashedPassword) {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
