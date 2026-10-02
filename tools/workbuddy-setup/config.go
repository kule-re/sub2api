package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type manifest struct {
	Schema int    `json:"schema_version"`
	Model  string `json:"model"`
	URL    string `json:"url"`
	APIKey string `json:"api_key"`
}

// Keep raw JSON values so unknown fields and encrypted credentials survive.
func mergeModels(raw []byte, m manifest) ([]byte, error) {
	var root map[string]json.RawMessage
	var models []json.RawMessage
	trimmed := bytes.TrimSpace(bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf}))
	wrapped := false
	if len(trimmed) > 0 {
		switch trimmed[0] {
		case '[':
			if err := json.Unmarshal(trimmed, &models); err != nil {
				return nil, errors.New("models.json is invalid; no changes made")
			}
		case '{':
			wrapped = true
			if err := json.Unmarshal(trimmed, &root); err != nil {
				return nil, err
			}
			list, ok := root["models"]
			if !ok || len(bytes.TrimSpace(list)) == 0 || bytes.TrimSpace(list)[0] != '[' {
				return nil, errors.New("unsupported models.json structure")
			}
			if err := json.Unmarshal(list, &models); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unsupported models.json structure")
		}
	}
	entry := map[string]any{"id": m.Model, "name": "Sub2API · " + m.Model, "vendor": "custom", "url": m.URL, "apiKey": m.APIKey,
		"supportsToolCall": true, "supportsImages": false, "useCustomProtocol": true}
	for _, item := range models {
		var old map[string]json.RawMessage
		if json.Unmarshal(item, &old) != nil || old == nil {
			return nil, errors.New("invalid existing model; no changes made")
		}
		var id string
		_ = json.Unmarshal(old["id"], &id)
		// Model IDs are sent upstream, so silently renaming a duplicate is unsafe.
		if id == m.Model {
			return nil, errors.New("this model ID already exists; edit it in WorkBuddy or choose another model")
		}
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	models = append(models, data)
	if !wrapped {
		return json.MarshalIndent(models, "", "  ")
	}
	root["models"], err = json.Marshal(models)
	if err != nil {
		return nil, err
	}
	if value, ok := root["availableModels"]; ok {
		var ids []string
		if json.Unmarshal(value, &ids) != nil {
			return nil, errors.New("unsupported availableModels structure")
		}
		found := false
		for _, id := range ids {
			if id == m.Model {
				found = true
			}
		}
		if !found {
			ids = append(ids, m.Model)
		}
		root["availableModels"], _ = json.Marshal(ids)
	}
	return json.MarshalIndent(root, "", "  ")
}

type restoreState struct {
	Existed bool   `json:"existed"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Backup  string `json:"backup"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func readConfig(path string) ([]byte, bool, error) {
	s, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !s.Mode().IsRegular() || s.Size() > 4<<20 {
		return nil, false, errors.New("configuration must be a regular file under 4 MiB")
	}
	b, err := os.ReadFile(path)
	return b, true, err
}

func replaceFile(path, stage string, data []byte) error {
	f, err := os.CreateTemp(stage, "write-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// Both the backup and the new file inherit the protected staging directory ACL.
func applyConfig(dir string, m manifest) error {
	return withConfigLock(dir, func(stage string) error {
		path := filepath.Join(dir, "models.json")
		old, exists, err := readConfig(path)
		if err != nil {
			return err
		}
		data, err := mergeModels(old, m)
		if err != nil {
			return err
		}
		backup := "models-" + time.Now().UTC().Format("20060102T150405.000000000") + ".json"
		if exists {
			if err = os.WriteFile(filepath.Join(stage, backup), old, 0600); err != nil {
				return err
			}
		}
		state := restoreState{Existed: exists, Before: digest(old), After: digest(data), Backup: backup}
		stateData, _ := json.Marshal(state)
		if err = replaceFile(filepath.Join(stage, "restore.json"), stage, stateData); err != nil {
			return err
		}
		current, currentExists, err := readConfig(path)
		if err != nil || currentExists != exists || !bytes.Equal(current, old) {
			return errors.New("configuration changed concurrently; retry after closing WorkBuddy")
		}
		return replaceFile(path, stage, data)
	})
}

func restoreConfig(dir string) error {
	return withConfigLock(dir, func(stage string) error {
		b, err := os.ReadFile(filepath.Join(stage, "restore.json"))
		if err != nil {
			return errors.New("no backup record found")
		}
		var state restoreState
		if json.Unmarshal(b, &state) != nil {
			return errors.New("invalid backup record")
		}
		current, exists, err := readConfig(filepath.Join(dir, "models.json"))
		if err != nil || !exists || digest(current) != state.After {
			return errors.New("configuration changed since setup; automatic restore refused to preserve newer edits")
		}
		if filepath.Base(state.Backup) != state.Backup {
			return errors.New("invalid backup path")
		}
		if !state.Existed {
			err = os.Remove(filepath.Join(dir, "models.json"))
		} else {
			old, readErr := os.ReadFile(filepath.Join(stage, state.Backup))
			if readErr != nil {
				return readErr
			}
			if digest(old) != state.Before {
				return errors.New("backup has changed; restore refused")
			}
			err = replaceFile(filepath.Join(dir, "models.json"), stage, old)
		}
		if err != nil {
			return err
		}
		return os.Remove(filepath.Join(stage, "restore.json"))
	})
}

func withConfigLock(dir string, fn func(string) error) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	// Refuse redirected directories; never operate through junctions/symlinks.
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if !samePath(resolved, abs) {
		return errors.New("redirected configuration directories are not supported")
	}
	stage := filepath.Join(dir, "sub2api-setup")
	if err = os.MkdirAll(stage, 0700); err != nil {
		return err
	}
	resolved, err = filepath.EvalSymlinks(stage)
	if err != nil || !samePath(resolved, stage) {
		return errors.New("redirected backup directory is not supported")
	}
	if err = protectDirectory(stage); err != nil {
		return fmt.Errorf("cannot protect backups: %w", err)
	}
	lock, err := os.OpenFile(filepath.Join(stage, "setup.lock"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("another setup is active (or a previous setup left setup.lock)")
	}
	defer os.Remove(lock.Name())
	defer lock.Close()
	return fn(stage)
}
