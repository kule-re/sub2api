package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const goodStream = "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"name\":\"sub2api_setup_check\",\"arguments\":\"{\\\"value\\\":\"}}]}}]}\n\n" +
	"data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"\\\"ready\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}\n\ndata: [DONE]\n\n"

func TestStreamingToolProbe(t *testing.T) {
	if err := checkStream(strings.NewReader(goodStream)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"data: [DONE]\n", strings.ReplaceAll(goodStream, "data: [DONE]", ""), strings.ReplaceAll(goodStream, "ready", "wrong"), `data: {"error":{"message":"fixture failure"}}`} {
		if checkStream(strings.NewReader(bad)) == nil {
			t.Fatal("accepted broken stream")
		}
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-only-not-a-key" {
			t.Error("missing bearer credential")
		}
		var payload map[string]any
		if json.NewDecoder(r.Body).Decode(&payload) != nil || payload["model"] != "new-model" || payload["stream"] != true || payload["tools"] == nil {
			t.Error("wrong probe payload")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(goodStream))
	}))
	defer server.Close()
	m := fixtureManifest()
	m.URL = server.URL
	if err := probe(server.Client(), m); err != nil {
		t.Fatal(err)
	}
}

func TestRedeemAndRedirectProtection(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/workbuddy/redeem" || r.URL.RawQuery != "" {
			t.Error("code exposed in URL or wrong route")
		}
		var body struct {
			Code string `json:"code"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Code == "redirect" {
			http.Redirect(w, r, "https://other.invalid", 307)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": fixtureManifest()})
	}))
	defer server.Close()
	c := server.Client()
	c.CheckRedirect = newClient().CheckRedirect
	if _, err := redeem(c, server.URL, "fixture-code"); err != nil {
		t.Fatal(err)
	}
	if _, err := redeem(c, server.URL, "redirect"); err == nil {
		t.Fatal("redirect accepted")
	}
	for _, u := range []string{"http://example.com", "https://u:p@example.com", "https://example.com?key=x", "https://example.com#key=x"} {
		if _, err := httpsURL(u); err == nil {
			t.Errorf("unsafe URL accepted: %s", u)
		}
	}
}
