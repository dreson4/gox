package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateIOS(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateIOS(IOSConfig{
		AppName:   "TestApp",
		BundleID:  "com.test.app",
		OutputDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("GenerateIOS failed: %v", err)
	}

	// Check core generated files exist
	expectedFiles := []string{
		"TestApp/bridge_core.m",
		"TestApp/main.m",
		"TestApp/Info.plist",
		"TestApp.xcodeproj/project.pbxproj",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Check component files exist
	componentFiles := []string{
		"TestApp/gox_view.m",
		"TestApp/gox_text.m",
		"TestApp/gox_button.m",
		"TestApp/gox_image.m",
		"TestApp/gox_textinput.m",
		"TestApp/gox_switch.m",
		"TestApp/gox_scrollview.m",
	}

	for _, f := range componentFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected component file %s to exist", f)
		}
	}

	// Check pbxproj contains bundle ID and all source files
	pbx, err := os.ReadFile(filepath.Join(tmpDir, "TestApp.xcodeproj/project.pbxproj"))
	if err != nil {
		t.Fatal(err)
	}
	pbxStr := string(pbx)
	if !strings.Contains(pbxStr, "com.test.app") {
		t.Error("pbxproj should contain bundle ID")
	}
	if !strings.Contains(pbxStr, "TestApp") {
		t.Error("pbxproj should contain app name")
	}
	if !strings.Contains(pbxStr, "bridge_core.m") {
		t.Error("pbxproj should reference bridge_core.m")
	}
	if !strings.Contains(pbxStr, "gox_text.m") {
		t.Error("pbxproj should reference component files")
	}
}

func TestGenerateIOSDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateIOS(IOSConfig{
		OutputDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("GenerateIOS with defaults failed: %v", err)
	}

	// Should use default name
	if _, err := os.Stat(filepath.Join(tmpDir, "GoxApp/bridge_core.m")); os.IsNotExist(err) {
		t.Error("should use default app name GoxApp")
	}
}
