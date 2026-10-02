package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	input := bufio.NewReader(os.Stdin)
	ask := func(label string) string {
		fmt.Print(label)
		line, _ := input.ReadString('\n')
		return strings.TrimSpace(line)
	}
	fmt.Println("Sub2API / WorkBuddy setup - Windows preview (WorkBuddy 5.6.2)")
	fmt.Println("No administrator permission needed. Close WorkBuddy before continuing.")
	if err := runSetup(ask); err != nil {
		fmt.Println("Setup stopped:", err)
	}
	ask("Press Enter to exit.")
}

func runSetup(ask func(string) string) error {
	if os.Getenv("WORKBUDDY_CONFIG_DIR") != "" || os.Getenv("CODEBUDDY_CONFIG_DIR") != "" {
		return errors.New("custom configuration directories are not supported in this preview")
	}
	exe, err := detectWorkBuddy()
	if err != nil {
		fmt.Println("Download and install the official WorkBuddy: https://www.codebuddy.cn/work/")
		exe = strings.Trim(ask("If already installed, enter the full path to WorkBuddy.exe (Enter to exit): "), "\"")
		if exe == "" {
			return errors.New("install WorkBuddy first, then run this helper again")
		}
	}
	if err = verifyVersion(exe); err != nil {
		return err
	}
	if err = ensureClosed(); err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".workbuddy")
	if strings.EqualFold(ask("Choose [setup / restore]: "), "restore") {
		if err = restoreConfig(dir); err != nil {
			return err
		}
		fmt.Println("Previous configuration restored.")
		return nil
	}
	site := ask("Website address (https://...): ")
	code := ask("Paste the one-time pairing code: ")
	m, err := redeem(newClient(), site, code)
	if err != nil {
		return err
	}
	fmt.Println("Model:", m.Model)
	fmt.Println("API endpoint:", m.URL)
	fmt.Println("A small billed streaming/tool-call test will run. Existing models will be preserved.")
	if !strings.EqualFold(ask("Continue? [yes/no]: "), "yes") {
		return errors.New("cancelled; no configuration written")
	}
	// Validate merge before spending tokens, and re-read before the actual write.
	old, _, err := readConfig(filepath.Join(dir, "models.json"))
	if err != nil {
		return err
	}
	if _, err = mergeModels(old, m); err != nil {
		return err
	}
	if err = probe(newClient(), m); err != nil {
		return err
	}
	if err = ensureClosed(); err != nil {
		return err
	}
	if err = applyConfig(dir, m); err != nil {
		return err
	}
	fmt.Println("Configuration saved. Backups:", filepath.Join(dir, "sub2api-setup"))
	fmt.Println("Open WorkBuddy, complete official sign-in if needed, and select Sub2API ·", m.Model)
	fmt.Println("Images are disabled. Real WorkBuddy task compatibility still needs verification.")
	if strings.EqualFold(ask("Open WorkBuddy now? [yes/no]: "), "yes") {
		return launchWorkBuddy(exe)
	}
	return nil
}
