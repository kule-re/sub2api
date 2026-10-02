package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func fixtureManifest() manifest {
	return manifest{Schema: 1, Model: "new-model", URL: "https://example.invalid/v1/chat/completions", APIKey: "test-only-not-a-key"}
}

func TestMergePreservesUnknownFieldsAndEncryptedKeys(t *testing.T) {
	raw := []byte(`{"models":[{"id":"old","apiKey":"WBEF1:opaque","extra":{"n":42}}],"availableModels":["old"],"unrelated":{"keep":true}}`)
	b, err := mergeModels(raw, fixtureManifest())
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Models          []map[string]any
		AvailableModels []string
		Unrelated       map[string]bool
	}
	if err = json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Models) != 2 || got.Models[0]["apiKey"] != "WBEF1:opaque" || got.Models[0]["extra"] == nil || !got.Unrelated["keep"] || len(got.AvailableModels) != 2 {
		t.Fatal("existing configuration lost")
	}
	if got.Models[1]["supportsImages"] != false || got.Models[1]["supportsToolCall"] != true {
		t.Fatal("invalid capability flags")
	}
	for _, bad := range []string{`[`, `null`, `{}`, `{"models":null}`, `[null]`, `[{"id":"new-model","url":"https://other.invalid"}]`, `{"models":[],"availableModels":42}`} {
		if _, err := mergeModels([]byte(bad), fixtureManifest()); err == nil {
			t.Errorf("accepted unsafe configuration %s", bad)
		}
	}
	b, err = mergeModels([]byte("\xef\xbb\xbf[]"), fixtureManifest())
	if err != nil || b[0] != '[' {
		t.Fatal("BOM or array format rejected")
	}
}

func TestApplyRestoreAndConcurrentChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.json")
	original := []byte(`{"models":[{"id":"other","apiKey":"fixture-only"}],"keep":99}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := applyConfig(dir, fixtureManifest()); err != nil {
		t.Fatal(err)
	}
	if err := restoreConfig(dir); err != nil {
		t.Fatal(err)
	}
	restored, _ := os.ReadFile(path)
	if !bytes.Equal(restored, original) {
		t.Fatal("restore not byte-for-byte")
	}
	if err := applyConfig(dir, fixtureManifest()); err != nil {
		t.Fatal(err)
	}
	edited := []byte(`[{"id":"user-edited-after-setup"}]`)
	if err := os.WriteFile(path, edited, 0600); err != nil {
		t.Fatal(err)
	}
	if err := restoreConfig(dir); err == nil {
		t.Fatal("restore overwrote newer edits")
	}
	b, _ := os.ReadFile(path)
	if !bytes.Equal(b, edited) {
		t.Fatal("newer edits changed")
	}
}

func TestNewConfigRestoreAndInvalidInput(t *testing.T) {
	dir := t.TempDir()
	if err := applyConfig(dir, fixtureManifest()); err != nil {
		t.Fatal(err)
	}
	if err := restoreConfig(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "models.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("new config was not removed")
	}
	bad := []byte("invalid JSON")
	if err := os.WriteFile(path, bad, 0600); err != nil {
		t.Fatal(err)
	}
	if err := applyConfig(dir, fixtureManifest()); err == nil {
		t.Fatal("invalid config overwritten")
	}
	b, _ := os.ReadFile(path)
	if !bytes.Equal(b, bad) {
		t.Fatal("invalid original changed")
	}
}
