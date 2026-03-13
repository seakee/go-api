package middleware

import "testing"

func TestIsSensitiveField_ExtendedKeys(t *testing.T) {
	sensitiveFields := []string{
		"password",
		"totp_key",
		"client_secret",
		"db_password",
		"credential_id",
		"client_data_json",
		"authenticator_data",
		"challenge_id",
	}

	for _, field := range sensitiveFields {
		if !isSensitiveField(field) {
			t.Fatalf("isSensitiveField(%q) = false, want true", field)
		}
	}
}

func TestIsSensitiveField_NonSensitiveKey(t *testing.T) {
	if isSensitiveField("display_name") {
		t.Fatalf("isSensitiveField(display_name) = true, want false")
	}
}
