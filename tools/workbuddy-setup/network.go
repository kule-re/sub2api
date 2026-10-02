package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func httpsURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("a plain HTTPS URL is required")
	}
	return u, nil
}

func newClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect refused") }}
}

func redeem(client *http.Client, site, code string) (manifest, error) {
	var result manifest
	u, err := httpsURL(site)
	if err != nil {
		return result, err
	}
	if u.Path != "" && u.Path != "/" {
		return result, errors.New("enter the website origin without a path")
	}
	u.Path = "/api/v1/workbuddy/redeem"
	body, _ := json.Marshal(map[string]string{"code": code})
	req, _ := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return result, errors.New("cannot reach the pairing service over HTTPS")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("pairing failed (HTTP %d); generate a new code on the website", resp.StatusCode)
	}
	var envelope struct {
		Code int      `json:"code"`
		Data manifest `json:"data"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 32<<10)).Decode(&envelope) != nil || envelope.Code != 0 {
		return result, errors.New("invalid pairing response")
	}
	result = envelope.Data
	if result.Schema != 1 || result.APIKey == "" || len(result.Model) == 0 || len(result.Model) > 200 {
		return manifest{}, errors.New("unsupported configuration manifest")
	}
	if _, err = httpsURL(result.URL); err != nil {
		return manifest{}, err
	}
	return result, nil
}

// A forced, tiny tool call validates auth, selected model, SSE and tool arguments.
// It does not execute any returned tool or claim image/task-level compatibility.
func probe(client *http.Client, m manifest) error {
	payload := map[string]any{"model": m.Model, "stream": true, "max_tokens": 128,
		"messages":    []any{map[string]string{"role": "user", "content": "Call sub2api_setup_check with value ready."}},
		"tools":       []any{map[string]any{"type": "function", "function": map[string]any{"name": "sub2api_setup_check", "description": "Connection test only", "parameters": map[string]any{"type": "object", "properties": map[string]any{"value": map[string]any{"type": "string", "enum": []string{"ready"}}}, "required": []string{"value"}, "additionalProperties": false}}}},
		"tool_choice": map[string]any{"type": "function", "function": map[string]string{"name": "sub2api_setup_check"}}}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, m.URL, bytes.NewReader(body))
	if err != nil {
		return errors.New("invalid model endpoint")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("model connection failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("model test failed (HTTP %d); no configuration written", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return errors.New("model did not return an SSE stream")
	}
	return checkStream(io.LimitReader(resp.Body, 1<<20))
}

func checkStream(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 128<<10)
	name, args := "", ""
	done := false
	finished := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			done = true
			break
		}
		var event struct {
			Error   json.RawMessage `json:"error"`
			Choices []struct {
				Index  int    `json:"index"`
				Finish string `json:"finish_reason"`
				Delta  struct {
					Tools []struct {
						Index    int `json:"index"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &event) != nil || len(event.Error) > 0 {
			return errors.New("invalid/error SSE response")
		}
		for _, choice := range event.Choices {
			if choice.Index != 0 {
				continue
			}
			if choice.Finish == "tool_calls" {
				finished = true
			}
			for _, call := range choice.Delta.Tools {
				if call.Index == 0 {
					name += call.Function.Name
					args += call.Function.Arguments
				}
			}
		}
	}
	var values struct {
		Value string `json:"value"`
	}
	if scanner.Err() != nil || !done || !finished || name != "sub2api_setup_check" || json.Unmarshal([]byte(args), &values) != nil || values.Value != "ready" {
		return errors.New("streaming tool-call test failed; no configuration written")
	}
	return nil
}
