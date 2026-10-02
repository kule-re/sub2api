//go:build !windows

package main

import (
	"errors"
	"os"
	"path/filepath"
)

func samePath(a, b string) bool         { return filepath.Clean(a) == filepath.Clean(b) }
func protectDirectory(dir string) error { return os.Chmod(dir, 0700) }
func detectWorkBuddy() (string, error)  { return "", errors.New("Windows only") }
func verifyVersion(string) error        { return errors.New("Windows only") }
func ensureClosed() error               { return errors.New("Windows only") }
func launchWorkBuddy(string) error      { return errors.New("Windows only") }
