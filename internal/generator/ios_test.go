package generator

import (
	"os"
	"path/filepath"
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

	// Check generated files exist
	expectedFiles := []string{
		"TestApp/bridge.m",
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

	// Check pbxproj contains bundle ID
	pbx, err := os.ReadFile(filepath.Join(tmpDir, "TestApp.xcodeproj/project.pbxproj"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(pbx), "com.test.app") {
		t.Error("pbxproj should contain bundle ID")
	}
	if !contains(string(pbx), "TestApp") {
		t.Error("pbxproj should contain app name")
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
	if _, err := os.Stat(filepath.Join(tmpDir, "GoxApp/bridge.m")); os.IsNotExist(err) {
		t.Error("should use default app name GoxApp")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
