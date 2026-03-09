package jwt

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost,
	)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(password),
	)

	return err == nil, err
}
