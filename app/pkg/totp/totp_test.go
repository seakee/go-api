package totp

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestNewGoogleAuthenticator(t *testing.T) {
	issuerName := "TestApp"
	config := NewGenerator(issuerName)

	if config.IssuerName != issuerName {
		t.Errorf("Expected issuer name %s, got %s", issuerName, config.IssuerName)
	}
	if config.HashAlgorithm != "SHA1" {
		t.Errorf("Expected hash algorithm SHA1, got %s", config.HashAlgorithm)
	}
	if config.UpdateInterval != 30 {
		t.Errorf("Expected update interval 30, got %d", config.UpdateInterval)
	}
	if config.CodeLength != 6 {
		t.Errorf("Expected code length 6, got %d", config.CodeLength)
	}
}

func TestGenerateQRCodeBase64(t *testing.T) {
	config := NewGenerator("TestApp")
	userAccount := "test@example.com"
	secretKey := "JBSWY3DPEHPK3PXP"

	qrCode, err := config.GenerateQRCodeBase64(userAccount, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate QR code: %v", err)
	}

	if !strings.HasPrefix(qrCode, "data:image/png;base64,") {
		t.Errorf("QR code does not have expected prefix")
	}

	// Decode base64 to ensure it's valid
	_, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(qrCode, "data:image/png;base64,"))
	if err != nil {
		t.Errorf("Generated QR code is not valid base64: %v", err)
	}
}

func TestGenerateTOTPCode(t *testing.T) {
	config := NewGenerator("TestApp")
	secretKey := "JBSWY3DPEHPK3PXP"

	// Test with a fixed time
	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	code := config.GenerateTOTPCode(secretKey, fixedTime)

	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	// Ensure the same time always generates the same code
	code2 := config.GenerateTOTPCode(secretKey, fixedTime)
	if code != code2 {
		t.Errorf("Codes do not match for the same time. Got %s and %s", code, code2)
	}

	// Ensure different times generate different codes
	laterTime := fixedTime.Add(31 * time.Second)
	laterCode := config.GenerateTOTPCode(secretKey, laterTime)
	if code == laterCode {
		t.Errorf("Codes should be different for different times")
	}
}

func TestVerifyTOTPCode(t *testing.T) {
	config := NewGenerator("TestApp")
	secretKey := "JBSWY3DPEHPK3PXP"

	// Generate a valid code
	validCode := config.GenerateTOTPCode(secretKey, time.Now())

	// Test valid code
	if !config.VerifyTOTPCode(validCode, secretKey, 1) {
		t.Errorf("Valid code was not verified")
	}

	// Test invalid code
	invalidCode := "000000"
	if config.VerifyTOTPCode(invalidCode, secretKey, 1) {
		t.Errorf("Invalid code was incorrectly verified")
	}

	// Test code from previous time window
	prevTime := time.Now().Add(-31 * time.Second)
	prevCode := config.GenerateTOTPCode(secretKey, prevTime)
	if !config.VerifyTOTPCode(prevCode, secretKey, 1) {
		t.Errorf("Code from previous time window was not verified")
	}

	// Test code from future time window
	futureTime := time.Now().Add(31 * time.Second)
	futureCode := config.GenerateTOTPCode(secretKey, futureTime)
	if !config.VerifyTOTPCode(futureCode, secretKey, 1) {
		t.Errorf("Code from future time window was not verified")
	}

	// Test code outside of allowed time window
	farFutureTime := time.Now().Add(90 * time.Second)
	farFutureCode := config.GenerateTOTPCode(secretKey, farFutureTime)
	if config.VerifyTOTPCode(farFutureCode, secretKey, 1) {
		t.Errorf("Code far outside time window was incorrectly verified")
	}
}
