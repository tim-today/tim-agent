package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/tim-today/tim-agent/pkg/appbundle"
	"github.com/tim-today/tim-agent/pkg/config"
	"github.com/tim-today/tim-agent/pkg/session"
	"github.com/tim-today/tim-agent/pkg/tray"
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

func TestMultiProjectManagementAndUniqueWorkDir(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.WorkDir = dirA
	cfg.Projects = []config.ProjectConfig{
		{
			ID:          "proj_a",
			Name:        "项目A",
			WorkDir:     config.NormalizeWorkDir(dirA),
			Agent:       "shell",
			AutoApprove: false,
		},
	}
	cfg.ActiveProjectID = "proj_a"

	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer session.GetManager().CloseAll()

	// 1. 新增第二个不同目录的项目 -> 应成功 (200)
	bodyB, _ := json.Marshal(map[string]any{
		"id":           "proj_b",
		"name":         "项目B",
		"work_dir":     dirB,
		"agent":        "shell",
		"auto_approve": true,
	})
	reqSaveB := httptest.NewRequest("POST", "/api/projects/save", bytes.NewReader(bodyB))
	reqSaveB.RemoteAddr = "127.0.0.1:54321"
	recSaveB := httptest.NewRecorder()
	s.handleAPIProjectSave(recSaveB, reqSaveB)
	if recSaveB.Code != http.StatusOK {
		t.Fatalf("新增不同目录项目B应成功(200)，实际: %d, body: %s", recSaveB.Code, recSaveB.Body.String())
	}

	// 2. 尝试新增第三个项目，但使用与项目A相同的目录 (包括末尾带斜杠) -> 必须拦截报错 (ok: false)
	bodyDup, _ := json.Marshal(map[string]any{
		"name":         "重复目录项目C",
		"work_dir":     dirA + "/",
		"agent":        "shell",
		"auto_approve": false,
	})
	reqSaveDup := httptest.NewRequest("POST", "/api/projects/save", bytes.NewReader(bodyDup))
	reqSaveDup.RemoteAddr = "127.0.0.1:54321"
	recSaveDup := httptest.NewRecorder()
	s.handleAPIProjectSave(recSaveDup, reqSaveDup)
	var dupResp map[string]any
	_ = json.Unmarshal(recSaveDup.Body.Bytes(), &dupResp)
	if dupResp["ok"] != false {
		t.Fatalf("相同目录属于不同项目必须被拦截(ok=false)，实际响应: %s", recSaveDup.Body.String())
	}

	// 3. 尝试把项目B的目录修改为项目A的目录 -> 同样必须被拦截 (ok: false)
	bodyEditDup, _ := json.Marshal(map[string]any{
		"id":           "proj_b",
		"name":         "项目B改目录",
		"work_dir":     dirA,
		"agent":        "shell",
		"auto_approve": true,
	})
	reqEditDup := httptest.NewRequest("POST", "/api/projects/save", bytes.NewReader(bodyEditDup))
	reqEditDup.RemoteAddr = "127.0.0.1:54321"
	recEditDup := httptest.NewRecorder()
	s.handleAPIProjectSave(recEditDup, reqEditDup)
	var editDupResp map[string]any
	_ = json.Unmarshal(recEditDup.Body.Bytes(), &editDupResp)
	if editDupResp["ok"] != false {
		t.Fatalf("修改项目目录与其他项目冲突时必须被拦截(ok=false)，实际响应: %s", recEditDup.Body.String())
	}

	// 4. 启动项目A会话，再切换到项目B会话 -> 验证两个会话在后台同时保持运行，且默认会话路由到最近活跃的项目B
	selectBodyA, _ := json.Marshal(map[string]any{"project_id": "proj_a"})
	reqSelectA := httptest.NewRequest("POST", "/api/projects/select", bytes.NewReader(selectBodyA))
	reqSelectA.RemoteAddr = "127.0.0.1:54321"
	recSelectA := httptest.NewRecorder()
	s.handleAPIProjectSelect(recSelectA, reqSelectA)
	if recSelectA.Code != http.StatusOK {
		t.Fatalf("启动并切换到项目A失败: %d, body: %s", recSelectA.Code, recSelectA.Body.String())
	}
	sessA := session.GetManager().FindByProjectOrDir("proj_a", dirA)
	if sessA == nil {
		t.Fatalf("项目A会话应当已被创建并运行")
	}

	selectBodyB, _ := json.Marshal(map[string]any{"id": "proj_b"})
	reqSelectB := httptest.NewRequest("POST", "/api/projects/select", bytes.NewReader(selectBodyB))
	reqSelectB.RemoteAddr = "127.0.0.1:54321"
	recSelectB := httptest.NewRecorder()
	s.handleAPIProjectSelect(recSelectB, reqSelectB)
	if recSelectB.Code != http.StatusOK {
		t.Fatalf("切换到项目B应成功(200)，实际: %d, body: %s", recSelectB.Code, recSelectB.Body.String())
	}

	// 验证项目A的后台会话依然存活，没有被关闭
	sessACheck := session.GetManager().FindByProjectOrDir("proj_a", dirA)
	if sessACheck == nil || sessACheck != sessA {
		t.Fatalf("切换到项目B后，项目A的会话应当继续在后台保持运行")
	}

	// 验证默认无参连接 ("default") 自动路由到最近操作的项目B
	targetProj := s.resolveTargetProject("default")
	if targetProj.ID != "proj_b" {
		t.Errorf("默认连接应路由到最近操作的项目 proj_b，实际路由到: %s", targetProj.ID)
	}

	// 模拟在项目A的会话中输入操作 (Touch)，验证最近活跃会话自动切回 proj_a
	sessA.Touch()
	targetProjAfterTouch := s.resolveTargetProject("default")
	if targetProjAfterTouch.ID != "proj_a" {
		t.Errorf("项目A产生交互操作后，默认连接应自动路由到最近操作的 proj_a，实际: %s", targetProjAfterTouch.ID)
	}
}

func TestGUIApplicationBundlePackaging(t *testing.T) {
	tmpDir := t.TempDir()
	dummyBin := filepath.Join(tmpDir, "dummy-bin")
	if err := os.WriteFile(dummyBin, []byte("#!/bin/sh\necho ok\n"), 0755); err != nil {
		t.Fatalf("写入测试二进制失败: %v", err)
	}

	// 1. 验证 ICNS 与 ICO 图标生成
	icns := tray.GenerateICNS()
	if len(icns) < 100 || string(icns[:4]) != "icns" {
		t.Fatalf("生成的 macOS .icns 格式不正确，长度: %d", len(icns))
	}
	ico := tray.GenerateICO()
	if len(ico) < 100 || ico[2] != 1 {
		t.Fatalf("生成的 Windows .ico 格式不正确，长度: %d", len(ico))
	}

	// 2. 验证 macOS TimAgent.app 应用程序包结构
	macAppDir := filepath.Join(tmpDir, "TimAgent.app")
	if err := appbundle.PackageMacApp(dummyBin, macAppDir); err != nil {
		t.Fatalf("PackageMacApp 失败: %v", err)
	}
	for _, rel := range []string{
		"Contents/Info.plist",
		"Contents/MacOS/tim-agent",
		"Contents/Resources/AppIcon.icns",
	} {
		if _, err := os.Stat(filepath.Join(macAppDir, rel)); err != nil {
			t.Errorf("macOS .app 缺少关键文件 %s: %v", rel, err)
		}
	}

	// 3. 验证 Linux 桌面 GUI 应用包结构
	linuxAppDir := filepath.Join(tmpDir, "TimAgent-linux-gui")
	if err := appbundle.PackageLinuxBundle(dummyBin, linuxAppDir); err != nil {
		t.Fatalf("PackageLinuxBundle 失败: %v", err)
	}
	for _, rel := range []string{
		"tim-agent",
		"TimAgent.desktop",
		"tim-agent.png",
		"Start-TimAgent-GUI.sh",
	} {
		if _, err := os.Stat(filepath.Join(linuxAppDir, rel)); err != nil {
			t.Errorf("Linux GUI 包缺少关键文件 %s: %v", rel, err)
		}
	}
}



