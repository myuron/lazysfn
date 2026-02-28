package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

func TestParseProfiles_Default(t *testing.T) {
	path := writeTempConfig(t, `[default]
region = ap-northeast-1
`)
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "default" {
		t.Errorf("expected profile name 'default', got %q", profiles[0].Name)
	}
}

func TestParseProfiles_NamedProfiles(t *testing.T) {
	path := writeTempConfig(t, `[profile dev]
region = us-east-1

[profile staging]
region = us-west-2

[profile prod]
region = ap-northeast-1
`)
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}

	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	expected := []string{"dev", "prod", "staging"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("profiles[%d].Name = %q, want %q", i, names[i], want)
		}
	}
}

func TestParseProfiles_MixedDefaultAndNamed(t *testing.T) {
	path := writeTempConfig(t, `[default]
region = ap-northeast-1

[profile dev]
region = us-east-1
`)
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	nameSet := make(map[string]bool)
	for _, p := range profiles {
		nameSet[p.Name] = true
	}
	if !nameSet["default"] {
		t.Error("expected 'default' profile to be present")
	}
	if !nameSet["dev"] {
		t.Error("expected 'dev' profile to be present")
	}
}

func TestParseProfiles_SortedByName(t *testing.T) {
	path := writeTempConfig(t, `[profile zebra]
region = us-east-1

[profile alpha]
region = us-west-2

[default]
region = ap-northeast-1

[profile middle]
region = eu-west-1
`)
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 4 {
		t.Fatalf("expected 4 profiles, got %d", len(profiles))
	}

	expected := []string{"alpha", "default", "middle", "zebra"}
	for i, want := range expected {
		if profiles[i].Name != want {
			t.Errorf("profiles[%d].Name = %q, want %q", i, profiles[i].Name, want)
		}
	}
}

func TestParseProfiles_WithRegion(t *testing.T) {
	path := writeTempConfig(t, `[default]
region = ap-northeast-1

[profile us-profile]
region = us-east-1
`)
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	regionMap := make(map[string]string)
	for _, p := range profiles {
		regionMap[p.Name] = p.Region
	}
	if regionMap["default"] != "ap-northeast-1" {
		t.Errorf("default region = %q, want %q", regionMap["default"], "ap-northeast-1")
	}
	if regionMap["us-profile"] != "us-east-1" {
		t.Errorf("us-profile region = %q, want %q", regionMap["us-profile"], "us-east-1")
	}
}

func TestParseProfiles_EmptyFile(t *testing.T) {
	path := writeTempConfig(t, "")
	profiles, err := ParseProfiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestParseProfiles_FileNotFound(t *testing.T) {
	_, err := ParseProfiles("/nonexistent/path/config")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseProfiles_MalformedConfig(t *testing.T) {
	path := writeTempConfig(t, `this is not a valid config file
=== garbage ===
[broken
no equals sign here
`)
	// Should not panic; may return error or empty list
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ParseProfiles panicked on malformed config: %v", r)
		}
	}()
	_, _ = ParseProfiles(path)
}
