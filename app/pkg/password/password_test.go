package password

import (
	"testing"

	"github.com/sk-pkg/util"
)

func TestHashAndVerifyBcryptCredential(t *testing.T) {
	credential := util.MD5("plain-password")

	hashed, err := HashCredential(credential)
	if err != nil {
		t.Fatalf("HashCredential() error = %v", err)
	}

	if !IsBcryptHash(hashed) {
		t.Fatalf("HashCredential() result is not bcrypt hash: %s", hashed)
	}

	matched, err := VerifyCredential(hashed, credential)
	if err != nil {
		t.Fatalf("VerifyCredential() error = %v", err)
	}

	if !matched {
		t.Fatalf("VerifyCredential() matched = false, want true")
	}
}

func TestVerifyCredential_LegacyHashNotSupported(t *testing.T) {
	credential := util.MD5("plain-password")
	legacySuffix := "legacy-marker"
	storedHash := util.MD5(credential + legacySuffix)

	matched, err := VerifyCredential(storedHash, credential)
	if err != nil {
		t.Fatalf("VerifyCredential() error = %v", err)
	}

	if matched {
		t.Fatalf("VerifyCredential() matched = true, want false")
	}
}

func TestHashCredential_EmptyInput(t *testing.T) {
	_, err := HashCredential("")
	if err == nil {
		t.Fatal("HashCredential(\"\") expected error, got nil")
	}
}

func TestVerifyCredential_EmptyStoredHash(t *testing.T) {
	matched, err := VerifyCredential("", "some-credential")
	if err != nil {
		t.Fatalf("VerifyCredential() error = %v", err)
	}
	if matched {
		t.Fatal("VerifyCredential() matched = true, want false")
	}
}

func TestVerifyCredential_NonBcryptHash(t *testing.T) {
	matched, err := VerifyCredential("not-a-bcrypt-hash", "some-credential")
	if err != nil {
		t.Fatalf("VerifyCredential() error = %v", err)
	}
	if matched {
		t.Fatal("VerifyCredential() matched = true, want false")
	}
}

func TestVerifyCredential_BcryptMismatch(t *testing.T) {
	hashed, err := HashCredential("correct-password")
	if err != nil {
		t.Fatalf("HashCredential() error = %v", err)
	}

	matched, err := VerifyCredential(hashed, "wrong-password")
	if err != nil {
		t.Fatalf("VerifyCredential() error = %v", err)
	}
	if matched {
		t.Fatal("VerifyCredential() matched = true, want false")
	}
}
