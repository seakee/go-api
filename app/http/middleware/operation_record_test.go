package middleware

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShouldOmitOperationPayload(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/go-api/internal/admin/auth/token", want: true},
		{path: "/go-api/internal/admin/auth/password/reset", want: true},
		{path: "/go-api/internal/admin/auth/passkey/register/finish", want: true},
		{path: "/go-api/internal/admin/auth/passkey/login/finish", want: true},
		{path: "/go-api/internal/admin/system/user/password/reset", want: true},
		{path: "/go-api/internal/admin/system/user/passkey", want: true},
		{path: "/go-api/internal/admin/system/user/paginate", want: false},
	}

	for _, tt := range tests {
		if got := shouldOmitOperationPayload(tt.path); got != tt.want {
			t.Errorf("shouldOmitOperationPayload(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestSanitizeQueryParams(t *testing.T) {
	raw := "name=alice&password=123456&token=abcdef&lang=zh-CN"

	got := sanitizeQueryParams(raw)
	var payload map[string]any
	if err := json.Unmarshal([]byte(got), &payload); err != nil {
		t.Fatalf("sanitizeQueryParams() json unmarshal error = %v, output=%s", err, got)
	}

	if payload["password"] != "[REDACTED]" {
		t.Fatalf("password not redacted: %v", payload["password"])
	}

	if payload["token"] != "[REDACTED]" {
		t.Fatalf("token not redacted: %v", payload["token"])
	}

	if payload["name"] != "alice" {
		t.Fatalf("name = %v, want alice", payload["name"])
	}
}

func TestSanitizeResponseBody(t *testing.T) {
	body := []byte(`{"code":0,"data":{"token":"abc","profile":{"password":"123","name":"alice"}}}`)
	got := sanitizeResponseBody(body)

	var payload map[string]any
	if err := json.Unmarshal([]byte(got), &payload); err != nil {
		t.Fatalf("sanitizeResponseBody() json unmarshal error = %v, output=%s", err, got)
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", payload["data"])
	}

	if data["token"] != "[REDACTED]" {
		t.Fatalf("token not redacted: %v", data["token"])
	}

	profile, ok := data["profile"].(map[string]any)
	if !ok {
		t.Fatalf("profile type = %T, want map[string]any", data["profile"])
	}

	if profile["password"] != "[REDACTED]" {
		t.Fatalf("password not redacted: %v", profile["password"])
	}
}

func TestSanitizeResponseBodyInvalidJSON(t *testing.T) {
	got := sanitizeResponseBody([]byte("not-json-response"))
	if !strings.Contains(got, "omitted response body") {
		t.Fatalf("sanitizeResponseBody() = %q, want omitted response body marker", got)
	}
}
