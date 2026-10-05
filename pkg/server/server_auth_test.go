package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tim-today/tim-agent/pkg/config"
)

func TestDashboardRemoteAccessRules(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Password = "testpwd123"
	cfg.AuthToken = "token123456"
	cfg.RemoteAccess = false

	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// 1. 本机请求 (127.0.0.1) 访问 / -> 免密放行返回 200
	reqLocal := httptest.NewRequest("GET", "/", nil)
	reqLocal.RemoteAddr = "127.0.0.1:54321"
	recLocal := httptest.NewRecorder()
	s.handleDashboard(recLocal, reqLocal)
	if recLocal.Code != http.StatusOK {
		t.Errorf("本地访问应直接放行(200)，实际状态码: %d", recLocal.Code)
	}

	// 2. 外部 IP 请求 (192.168.1.100)，未开启远程访问 -> 302 重定向至 /terminal
	reqRemoteBlocked := httptest.NewRequest("GET", "/", nil)
	reqRemoteBlocked.RemoteAddr = "192.168.1.100:54321"
	recRemoteBlocked := httptest.NewRecorder()
	s.handleDashboard(recRemoteBlocked, reqRemoteBlocked)
	if recRemoteBlocked.Code != http.StatusFound || recRemoteBlocked.Header().Get("Location") != "/terminal" {
		t.Errorf("未开启远程访问时应重定向到 /terminal，实际: %d -> %s", recRemoteBlocked.Code, recRemoteBlocked.Header().Get("Location"))
	}

	// 3. 开启远程访问，外部 IP 未认证 -> 302 重定向至 /login?redirect=/
	cfg.RemoteAccess = true
	reqRemoteUnauth := httptest.NewRequest("GET", "/", nil)
	reqRemoteUnauth.RemoteAddr = "192.168.1.100:54321"
	recRemoteUnauth := httptest.NewRecorder()
	s.handleDashboard(recRemoteUnauth, reqRemoteUnauth)
	if recRemoteUnauth.Code != http.StatusFound || recRemoteUnauth.Header().Get("Location") != "/login?redirect=/" {
		t.Errorf("开启远程访问但未认证应重定向到 /login?redirect=/，实际: %d -> %s", recRemoteUnauth.Code, recRemoteUnauth.Header().Get("Location"))
	}

	// 4. 开启远程访问，外部 IP 带合法 Token -> 200 放行
	reqRemoteAuthed := httptest.NewRequest("GET", "/?token=token123456", nil)
	reqRemoteAuthed.RemoteAddr = "192.168.1.100:54321"
	recRemoteAuthed := httptest.NewRecorder()
	s.handleDashboard(recRemoteAuthed, reqRemoteAuthed)
	if recRemoteAuthed.Code != http.StatusOK {
		t.Errorf("认证通过后应返回 200，实际状态码: %d", recRemoteAuthed.Code)
	}

	// 5. 外部 API 调用 /api/config 权限测试
	cfg.RemoteAccess = false
	payload, _ := json.Marshal(map[string]any{"port": 20996, "language": "en"})
	reqAPIBlocked := httptest.NewRequest("POST", "/api/config", bytes.NewReader(payload))
	reqAPIBlocked.RemoteAddr = "192.168.1.100:54321"
	recAPIBlocked := httptest.NewRecorder()
	s.handleAPIConfig(recAPIBlocked, reqAPIBlocked)
	if recAPIBlocked.Code != http.StatusForbidden {
		t.Errorf("未开启远程访问调用配置API应返回 403，实际: %d", recAPIBlocked.Code)
	}

	// 6. /api/status 外部未认证访问脱敏
	reqStatus := httptest.NewRequest("GET", "/api/status", nil)
	reqStatus.RemoteAddr = "192.168.1.100:54321"
	recStatus := httptest.NewRecorder()
	s.handleAPIStatus(recStatus, reqStatus)
	var statusData map[string]any
	_ = json.Unmarshal(recStatus.Body.Bytes(), &statusData)
	if statusData["auth_token"] != "" {
		t.Errorf("未经认证的外部状态请求不应暴露 auth_token，实际: %v", statusData["auth_token"])
	}
}

func TestPWAResourcesAndHeaders(t *testing.T) {
	cfg := config.DefaultConfig()
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// 1. 测试 /manifest.json
	reqManifest := httptest.NewRequest("GET", "/manifest.json", nil)
	recManifest := httptest.NewRecorder()
	s.handleManifest(recManifest, reqManifest)
	if recManifest.Code != http.StatusOK {
		t.Fatalf("manifest.json 应返回 200，实际: %d", recManifest.Code)
	}
	var manifestMap map[string]any
	if err := json.Unmarshal(recManifest.Body.Bytes(), &manifestMap); err != nil {
		t.Fatalf("manifest.json 格式应为有效 JSON: %v", err)
	}
	if manifestMap["start_url"] != "/terminal" {
		t.Errorf("manifest start_url 应为 /terminal，实际: %v", manifestMap["start_url"])
	}

	// 2. 测试 /sw.js (核心：严格禁用 HTTP 缓存，防止死锁旧版本)
	reqSW := httptest.NewRequest("GET", "/sw.js", nil)
	recSW := httptest.NewRecorder()
	s.handleServiceWorker(recSW, reqSW)
	if recSW.Code != http.StatusOK {
		t.Fatalf("sw.js 应返回 200，实际: %d", recSW.Code)
	}
	cacheControl := recSW.Header().Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("sw.js 必须具备 no-cache, no-store 响应头避免顽固缓存，实际: %s", cacheControl)
	}
	contentType := recSW.Header().Get("Content-Type")
	if contentType != "application/javascript; charset=utf-8" {
		t.Errorf("sw.js Content-Type 应为 application/javascript，实际: %s", contentType)
	}

	// 3. 测试 /favicon.ico
	reqFavicon := httptest.NewRequest("GET", "/favicon.ico", nil)
	recFavicon := httptest.NewRecorder()
	s.handleFavicon(recFavicon, reqFavicon)
	if recFavicon.Code != http.StatusOK {
		t.Fatalf("favicon.ico 应返回 200，实际: %d", recFavicon.Code)
	}
}

