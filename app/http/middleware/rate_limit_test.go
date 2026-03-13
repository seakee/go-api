package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/config"
	"github.com/sk-pkg/redis"
)

func TestAdminAuthRateLimit_DisabledPassThrough(t *testing.T) {
	loadRateLimitTestConfig(t, false)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware{}.AdminAuthRateLimit())
	r.POST("/go-api/internal/admin/auth/token", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-api/internal/admin/auth/token", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestAdminAuthRateLimit_EnabledWithoutRedisPassThrough(t *testing.T) {
	loadRateLimitTestConfig(t, true)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware{}.AdminAuthRateLimit())
	r.POST("/go-api/internal/admin/auth/token", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-api/internal/admin/auth/token", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestAdminAuthRateLimit_ThrottleAfterLimit(t *testing.T) {
	loadRateLimitTestConfig(t, true)

	originalEval := adminAuthRateLimitEval
	t.Cleanup(func() {
		adminAuthRateLimitEval = originalEval
	})

	var (
		evalCalls   int
		receivedKey string
		receivedTTL int
	)

	adminAuthRateLimitEval = func(ctx context.Context, manager *redis.Manager, key string, window int) (int64, error) {
		evalCalls++
		receivedKey = key
		receivedTTL = window
		return int64(evalCalls), nil
	}

	manager := &redis.Manager{}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware{redis: map[string]*redis.Manager{"go-api": manager}}.AdminAuthRateLimit())
	r.POST("/go-api/internal/admin/auth/token", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 1; i <= 20; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/go-api/internal/admin/auth/token", nil)
		req.RemoteAddr = "192.0.2.10:12345"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i, w.Code, http.StatusNoContent)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-api/internal/admin/auth/token", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if evalCalls != 21 {
		t.Fatalf("evalCalls = %d, want 21", evalCalls)
	}
	if receivedTTL != 60 {
		t.Fatalf("receivedTTL = %d, want 60", receivedTTL)
	}
	// LuaWithContext 不自动加前缀，key 中包含 manager.Prefix（此处为空字符串）
	if receivedKey != "admin:auth:rate-limit:/go-api/internal/admin/auth/token:192.0.2.10" {
		t.Fatalf("receivedKey = %s, want %s", receivedKey, "admin:auth:rate-limit:/go-api/internal/admin/auth/token:192.0.2.10")
	}
}

func TestAdminAuthRateLimit_PasskeyLoginPath(t *testing.T) {
	loadRateLimitTestConfig(t, true)

	originalEval := adminAuthRateLimitEval
	t.Cleanup(func() {
		adminAuthRateLimitEval = originalEval
	})

	receivedKey := ""
	adminAuthRateLimitEval = func(ctx context.Context, manager *redis.Manager, key string, window int) (int64, error) {
		receivedKey = key
		return 1, nil
	}

	manager := &redis.Manager{}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware{redis: map[string]*redis.Manager{"go-api": manager}}.AdminAuthRateLimit())
	r.POST("/go-api/internal/admin/auth/passkey/login/options", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-api/internal/admin/auth/passkey/login/options", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if receivedKey != "admin:auth:rate-limit:/go-api/internal/admin/auth/passkey/login/options:192.0.2.10" {
		t.Fatalf("receivedKey = %s, want %s", receivedKey, "admin:auth:rate-limit:/go-api/internal/admin/auth/passkey/login/options:192.0.2.10")
	}
}

func loadRateLimitTestConfig(t *testing.T, enabled bool) {
	t.Helper()

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "bin", "configs")
	if err = os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir error: %v", err)
	}

	cfg := fmt.Sprintf(`{
  "system": {
    "name": "go-api-test",
    "run_mode": "debug",
    "http_port": ":8080",
    "read_timeout": 60,
    "write_timeout": 60,
    "version": "1.0.0",
    "debug_mode": true,
    "default_lang": "zh-CN",
    "jwt_secret": "unit-test-secret",
    "token_expire": 3600,
    "admin": {
      "auth_rate_limit": {
        "enable": %t,
        "window_seconds": 60,
        "max_requests": 20
      }
    }
  }
}`, enabled)

	configFile := filepath.Join(configDir, "test.json")
	if err = os.WriteFile(configFile, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write test config error: %v", err)
	}

	if err = os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	t.Setenv("RUN_ENV", "test")

	if _, err = config.LoadConfig(); err != nil {
		t.Fatalf("load config error: %v", err)
	}
}
