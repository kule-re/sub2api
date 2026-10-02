package service

import "strings"

const (
	defaultSiteName = "哈呀哈基米ai"
	defaultSiteLogo = "/hakimi-logo.png"
)

// Retain administrator-defined brands while replacing previous defaults in
// existing installations as well as newly initialized settings.
func normalizeSiteName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "Sub2API" || value == "悟狗ai" {
		return defaultSiteName
	}
	return value
}

func normalizeSiteLogo(value string) string {
	value = strings.TrimSpace(value)
	path, _, _ := strings.Cut(value, "?")
	path, _, _ = strings.Cut(path, "#")
	if value == "" || path == "/logo.svg" || path == "logo.svg" || path == "./logo.svg" || path == "/wugou-logo.png" || path == "wugou-logo.png" || path == "./wugou-logo.png" {
		return defaultSiteLogo
	}
	return value
}
