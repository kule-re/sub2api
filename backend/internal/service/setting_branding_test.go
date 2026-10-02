package service

import "testing"

func TestSiteBrandingExistingDefaults(t *testing.T) {
	for _, name := range []string{"", "  ", "Sub2API", " Sub2API ", "悟狗ai", " 悟狗ai ", "哈呀哈基米ai"} {
		if got := normalizeSiteName(name); got != "哈呀哈基米ai" {
			t.Errorf("normalizeSiteName(%q) = %q", name, got)
		}
	}
	for _, logo := range []string{"", "  ", "/logo.svg", "logo.svg", "./logo.svg", "/logo.svg?v=1#icon", "/wugou-logo.png", "wugou-logo.png", "./wugou-logo.png", " /wugou-logo.png?v=1#icon ", "/wugou-logo.png#icon", "/hakimi-logo.png"} {
		if got := normalizeSiteLogo(logo); got != "/hakimi-logo.png" {
			t.Errorf("normalizeSiteLogo(%q) = %q", logo, got)
		}
	}
}

func TestSiteBrandingCustomSettingsRemain(t *testing.T) {
	for _, name := range []string{"Custom Site", "悟狗ai定制", "哈呀哈基米ai"} {
		if got := normalizeSiteName(name); got != name {
			t.Errorf("custom site name changed: %q", got)
		}
	}
	for _, logo := range []string{"/custom.svg", "https://example.com/logo.svg?v=1", "https://example.com/wugou-logo.png?v=1", "/custom/wugou-logo.png", "/hakimi-logo.png?v=2#icon", "data:image/png;base64,abc"} {
		if got := normalizeSiteLogo(logo); got != logo {
			t.Errorf("custom site logo changed: %q", got)
		}
	}
}
