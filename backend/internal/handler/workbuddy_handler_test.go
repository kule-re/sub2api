package handler

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/workbuddypair"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type workBuddyFakeKeys struct {
	key      *service.APIKey
	disabled bool
}

func TestWorkBuddyIssueRequiresOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, loggedIn := range []bool{false, true} {
		h := &WorkBuddyHandler{keys: &workBuddyFakeKeys{key: &service.APIKey{ID: 1, UserID: 99, Status: "active", Group: &service.Group{ID: 3, Platform: "openai", Status: "active"}}}}
		router := gin.New()
		router.POST("/pair", func(c *gin.Context) {
			if loggedIn {
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
			}
			h.Issue(c)
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/pair", strings.NewReader(`{"key_id":1,"model":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		want := 401
		if loggedIn {
			want = 400
		}
		if w.Code != want {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

func (f *workBuddyFakeKeys) GetByID(context.Context, int64) (*service.APIKey, error) {
	return f.key, nil
}
func (f *workBuddyFakeKeys) ValidateKey(context.Context, string) (*service.APIKey, *service.User, error) {
	if f.disabled {
		return nil, nil, errors.New("disabled")
	}
	return f.key, &service.User{ID: 7, Status: "active"}, nil
}

func TestWorkBuddyRedeemRevalidatesAndConsumes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []string{"valid", "owner-changed", "revoked", "group-changed", "expired", "user-disabled"} {
		t.Run(scenario, func(t *testing.T) {
			r := miniredis.RunT(t)
			client := redis.NewClient(&redis.Options{Addr: r.Addr()})
			defer client.Close()
			keys := &workBuddyFakeKeys{key: &service.APIKey{ID: 1, UserID: 7, Key: "fixture-key", Status: "active", Group: &service.Group{ID: 3, Platform: "openai", Status: "active"}}}
			h := &WorkBuddyHandler{keys: keys, store: workbuddypair.Store{Client: client}}
			token, err := h.store.Issue(context.Background(), workbuddypair.Grant{KeyID: 1, UserID: 7, GroupID: 3, Model: "test", URL: "https://example.invalid/v1/chat/completions"})
			if err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "owner-changed":
				keys.key.UserID = 99
			case "revoked":
				keys.key.Status = "inactive"
			case "group-changed":
				keys.key.Group.ID = 8
			case "expired":
				past := time.Now().Add(-time.Second)
				keys.key.ExpiresAt = &past
			case "user-disabled":
				keys.disabled = true
			}
			router := gin.New()
			router.POST("/redeem", h.Redeem)
			request := func() *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/redeem", strings.NewReader(`{"code":"`+token+`"}`))
				req.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, req)
				return w
			}
			w := request()
			if scenario == "valid" {
				if w.Code != 200 || !strings.Contains(w.Body.String(), "fixture-key") {
					t.Fatal("valid redemption failed")
				}
			} else if w.Code != 403 || strings.Contains(w.Body.String(), "fixture-key") {
				t.Fatal("invalid credential released")
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("credential response can be cached")
			}
			if request().Code != 400 {
				t.Fatal("pairing code reused")
			}
		})
	}
}

func TestWorkBuddyEndpoint(t *testing.T) {
	for _, base := range []string{"https://api.example.com", "https://api.example.com/v1/", "https://api.example.com/v1/chat/completions"} {
		u, err := workBuddyEndpoint(base)
		if err != nil || u != "https://api.example.com/v1/chat/completions" {
			t.Fatal("incorrect endpoint", u, err)
		}
	}
	for _, base := range []string{"", "http://example.com", "https://u:p@example.com", "https://example.com?secret=x"} {
		if _, err := workBuddyEndpoint(base); err == nil {
			t.Fatal("unsafe endpoint accepted")
		}
	}
}
