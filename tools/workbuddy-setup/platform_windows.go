package main

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func samePath(a, b string) bool { return strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) }
func command(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return c
}

func protectDirectory(dir string) error {
	// Replace the DACL, including explicit grants from previous runs. Pass the
	// path as data in an environment variable, never interpolate it into code.
	c := command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", `
$ErrorActionPreference = 'Stop'
$acl = [System.IO.Directory]::GetAccessControl($env:SUB2API_SETUP_ACL_PATH, [System.Security.AccessControl.AccessControlSections]::Access)
$acl.SetAccessRuleProtection($true, $false)
$owner = [System.Security.Principal.WindowsIdentity]::GetCurrent().User
foreach ($existing in @($acl.Access)) { $acl.RemoveAccessRuleSpecific($existing) }
$system = New-Object System.Security.Principal.SecurityIdentifier('S-1-5-18')
foreach ($sid in @($owner, $system)) {
  $rule = New-Object System.Security.AccessControl.FileSystemAccessRule($sid, 'FullControl', 'ContainerInherit,ObjectInherit', 'None', 'Allow')
  $acl.AddAccessRule($rule)
}
[System.IO.Directory]::SetAccessControl($env:SUB2API_SETUP_ACL_PATH, $acl)
`)
	c.Env = append(os.Environ(), "SUB2API_SETUP_ACL_PATH="+dir)
	output, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Windows ACL update failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func detectWorkBuddy() (string, error) {
	for _, base := range []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs"), os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		p := filepath.Join(base, "WorkBuddy", "WorkBuddy.exe")
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("WorkBuddy not found")
}

// Read only package metadata from Electron ASAR, never local user credentials.
func verifyVersion(exe string) error {
	if !strings.EqualFold(filepath.Base(exe), "WorkBuddy.exe") {
		return errors.New("select the domestic WorkBuddy.exe; other editions are not validated")
	}
	if s, err := os.Stat(exe); err != nil || !s.Mode().IsRegular() {
		return errors.New("WorkBuddy.exe not found")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(exe), "resources", "app.asar"))
	if err != nil {
		return err
	}
	defer f.Close()
	var b [16]byte
	if _, err = io.ReadFull(f, b[:]); err != nil {
		return err
	}
	n := binary.LittleEndian.Uint32(b[12:])
	if n == 0 || n > 8<<20 {
		return errors.New("unsupported WorkBuddy package")
	}
	h := make([]byte, n)
	if _, err = io.ReadFull(f, h); err != nil {
		return err
	}
	var header struct {
		Files map[string]struct {
			Size   int64  `json:"size"`
			Offset string `json:"offset"`
		} `json:"files"`
	}
	if json.Unmarshal(h, &header) != nil {
		return errors.New("invalid package metadata")
	}
	entry, ok := header.Files["package.json"]
	if !ok || entry.Size <= 0 || entry.Size > 64<<10 {
		return errors.New("missing package metadata")
	}
	offset, err := strconv.ParseInt(entry.Offset, 10, 64)
	if err != nil || offset < 0 {
		return errors.New("invalid package offset")
	}
	if _, err = f.Seek(int64(binary.LittleEndian.Uint32(b[4:]))+8+offset, io.SeekStart); err != nil {
		return err
	}
	data := make([]byte, entry.Size)
	if _, err = io.ReadFull(f, data); err != nil {
		return err
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &pkg) != nil || pkg.Version != "5.6.2" {
		return errors.New("only WorkBuddy 5.6.2 is validated; use its Settings > Models for other versions")
	}
	return nil
}

func ensureClosed() error {
	b, err := command("tasklist.exe", "/fo", "csv", "/nh").Output()
	if err != nil {
		return errors.New("cannot check running applications")
	}
	rows, err := csv.NewReader(strings.NewReader(string(b))).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range rows {
		if len(row) > 0 && (strings.EqualFold(row[0], "WorkBuddy.exe") || strings.EqualFold(row[0], "WorkBuddyAI.exe")) {
			return errors.New("quit WorkBuddy from the system tray, then run setup again")
		}
	}
	return nil
}
func launchWorkBuddy(path string) error { return exec.Command(path).Start() }
