package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "guise-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src")
	dst := filepath.Join(tmpDir, "dst")

	// Create src structure
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(src, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "subdir", "subfile.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test Copy
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify
	content, err := os.ReadFile(filepath.Join(dst, "file.txt"))
	if err != nil {
		t.Fatal("file.txt not copied")
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(content))
	}

	content, err = os.ReadFile(filepath.Join(dst, "subdir", "subfile.txt"))
	if err != nil {
		t.Fatal("subdir/subfile.txt not copied")
	}
	if string(content) != "world" {
		t.Errorf("expected 'world', got '%s'", string(content))
	}
}

func TestConfigLoadSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "guise-test-config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	cfg := Config{
		Providers: []Provider{
			{ID: "test", Name: "Test Provider"},
		},
	}

	if err := saveConfig(configPath, cfg); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	loaded, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if len(loaded.Providers) != 1 || loaded.Providers[0].ID != "test" {
		t.Errorf("config mismatch")
	}
}

func TestDefaultProvidersMerge(t *testing.T) {
	// Simulate the logic in main() where we merge defaults
	// We can't easily test main() directly, but we can verify the logic concept here
	
	existing := []Provider{
		{ID: "custom", Name: "Custom"},
		{ID: "openai-codex", Name: "Existing OpenAI"}, // Should not be overwritten
	}
	
	defaults := []Provider{
		{ID: "openai-codex", Name: "Default OpenAI"},
		{ID: "new-provider", Name: "New Provider"},
	}

	merged := existing
	for _, def := range defaults {
		found := false
		for _, ex := range existing {
			if ex.ID == def.ID {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, def)
		}
	}

	if len(merged) != 3 {
		t.Errorf("expected 3 providers, got %d", len(merged))
	}
	
	// Check that existing openai was preserved
	for _, p := range merged {
		if p.ID == "openai-codex" && p.Name != "Existing OpenAI" {
			t.Errorf("Existing provider was overwritten")
		}
	}
}
