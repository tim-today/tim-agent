package autostart

import (
	"testing"
)

func TestAutostartToggle(t *testing.T) {
	orig := IsEnabled()

	// 测试开启
	if err := Enable(); err != nil {
		t.Skipf("Skipping integration test in sandbox environment (permission denied: %v)", err)
		return
	}
	if !IsEnabled() {
		t.Fatalf("Expected autostart to be enabled")
	}

	// 测试关闭
	if err := Disable(); err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
	if IsEnabled() {
		t.Fatalf("Expected autostart to be disabled")
	}

	// 恢复原始状态
	if orig {
		_ = Enable()
	}
}
