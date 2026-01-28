// Package totp provides Time-based One-Time Password (TOTP) generation and verification.
package totp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image/png"
	"net/url"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// Generator represents a Time-based One-Time Password (TOTP) configuration.
type Generator struct {
	IssuerName     string // Issuer name, typically the application or organization name
	HashAlgorithm  string // Hash algorithm used, defaults to SHA1
	UpdateInterval int64  // TOTP code update interval in seconds
	CodeLength     int    // Length of the generated verification code
}

// GenerateQRCodeBase64 generates a QR code containing TOTP information and returns a base64-encoded PNG image.
//
// Parameters:
//   - userAccount: User account name.
//   - secretKey: User's secret key.
//
// Returns:
//   - string: Base64-encoded PNG image string.
//   - error: Error if QR code generation fails.
func (g Generator) GenerateQRCodeBase64(userAccount, secretKey string) (string, error) {
	totpURL := fmt.Sprintf(
		"otpauth://totp/%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		url.QueryEscape(userAccount),
		secretKey,
		url.QueryEscape(g.IssuerName),
		g.HashAlgorithm,
		g.CodeLength,
		g.UpdateInterval,
	)

	// Generate QR code
	qrCode, err := qr.Encode(totpURL, qr.M, qr.Auto)
	if err != nil {
		return "", err
	}

	// Scale QR code
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		return "", err
	}

	// Encode QR code to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, qrCode)
	if err != nil {
		return "", err
	}

	// Encode PNG to base64
	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Return complete base64-encoded image string, can be used directly in HTML img tag
	return "data:image/png;base64," + base64Image, nil
}

// GenerateTOTPCode generates a TOTP verification code for the specified time.
//
// Parameters:
//   - secretKey: User's secret key.
//   - t: Time for which to generate the code.
//
// Returns:
//   - string: Generated TOTP code.
func (g Generator) GenerateTOTPCode(secretKey string, t time.Time) string {
	decodedKey, err := base32.StdEncoding.DecodeString(strings.ToUpper(secretKey))
	if err != nil {
		return ""
	}

	timeInterval := t.Unix() / g.UpdateInterval
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeInterval))

	hmacHash := hmac.New(sha1.New, decodedKey)
	hmacHash.Write(timeBytes)
	hashResult := hmacHash.Sum(nil)

	offset := hashResult[len(hashResult)-1] & 0xF
	truncatedHash := hashResult[offset : offset+4]

	totpValue := binary.BigEndian.Uint32(truncatedHash) & 0x7FFFFFFF
	totpCode := totpValue % uint32(powerOfTen(g.CodeLength))

	return fmt.Sprintf("%0*d", g.CodeLength, totpCode)
}

// VerifyTOTPCode verifies whether the provided TOTP code is correct.
//
// Parameters:
//   - userProvidedCode: TOTP code provided by the user.
//   - secretKey: User's secret key.
//   - timeWindow: Allowed time window (number of intervals before and after).
//
// Returns:
//   - bool: Whether the verification passed.
func (g Generator) VerifyTOTPCode(userProvidedCode, secretKey string, timeWindow int) bool {
	if len(userProvidedCode) != g.CodeLength {
		return false
	}

	currentTime := time.Now()
	for i := -timeWindow; i <= timeWindow; i++ {
		checkTime := currentTime.Add(time.Duration(i) * time.Duration(g.UpdateInterval) * time.Second)
		if g.GenerateTOTPCode(secretKey, checkTime) == userProvidedCode {
			return true
		}
	}
	return false
}

// NewGenerator creates a new Generator instance with default settings.
//
// Parameters:
//   - issuerName: Issuer name for the TOTP.
//
// Returns:
//   - Generator: New Generator instance.
func NewGenerator(issuerName string) Generator {
	return Generator{
		IssuerName:     issuerName,
		HashAlgorithm:  "SHA1",
		UpdateInterval: 30,
		CodeLength:     6,
	}
}

// powerOfTen calculates 10 to the power of n.
func powerOfTen(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
