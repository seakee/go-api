package password

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptPrefix2A = "$2a$"
	bcryptPrefix2B = "$2b$"
	bcryptPrefix2Y = "$2y$"
	bcryptCost     = 12
)

// IsBcryptHash reports whether value looks like a bcrypt hash string.
func IsBcryptHash(value string) bool {
	return strings.HasPrefix(value, bcryptPrefix2A) ||
		strings.HasPrefix(value, bcryptPrefix2B) ||
		strings.HasPrefix(value, bcryptPrefix2Y)
}

// HashCredential hashes a credential digest using bcrypt.
func HashCredential(credential string) (string, error) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return "", errors.New("credential is empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(credential), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// VerifyCredential verifies credential against a stored hash.
func VerifyCredential(storedHash, credential string) (matched bool, err error) {
	if storedHash == "" || credential == "" {
		return false, nil
	}

	if !IsBcryptHash(storedHash) {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(credential))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}
