package passhash

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func NewPasswordHasherBcrypt(cost int) *PasswordHasherBcrypt {
	switch {
	case cost == 0:
		cost = bcrypt.DefaultCost
	case cost < bcrypt.MinCost:
		cost = bcrypt.MinCost
	case cost > bcrypt.MaxCost:
		cost = bcrypt.MaxCost
	}
	return &PasswordHasherBcrypt{cost: cost}
}

// PasswordHasherBcrypt hashes passwords using scrypt
type PasswordHasherBcrypt struct{ cost int }

func (h *PasswordHasherBcrypt) HashPassword(plainText []byte) (hash string, err error) {
	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	b, err := bcrypt.GenerateFromPassword(plainText, bcrypt.DefaultCost)
	return string(b), err
}

func (h *PasswordHasherBcrypt) ComparePassword(
	plainText []byte, hash []byte,
) (ok bool, err error) {
	err = bcrypt.CompareHashAndPassword(hash, plainText)
	return err == nil, err
}
