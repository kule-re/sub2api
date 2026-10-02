package handler

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/workbuddypair"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type workBuddyKeyReader interface {
	GetByID(context.Context, int64) (*service.APIKey, error)
	ValidateKey(context.Context, string) (*service.APIKey, *service.User, error)
}

type WorkBuddyHandler struct {
	keys     workBuddyKeyReader
	settings *service.SettingService
	store    workbuddypair.Store
}

func NewWorkBuddyHandler(keys *service.APIKeyService, settings *service.SettingService, client *redis.Client) *WorkBuddyHandler {
	return &WorkBuddyHandler{keys: keys, settings: settings, store: workbuddypair.Store{Client: client}}
}

var workBuddyModelID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:/-]{0,199}$`)

func workBuddyEndpoint(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", workbuddypair.ErrInvalid
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(u.Path, "/chat/completions") {
		if !strings.HasSuffix(u.Path, "/v1") {
			u.Path += "/v1"
		}
		u.Path += "/chat/completions"
	}
	u.RawPath = ""
	return u.String(), nil
}

func eligibleWorkBuddyKey(key *service.APIKey, owner int64) bool {
	return key != nil && key.UserID == owner && key.IsActive() && !key.IsExpired() && !key.IsQuotaExhausted() &&
		key.Group != nil && key.Group.Platform == "openai" && key.Group.Status == "active"
}

// Issue requires the existing panel JWT; no secret is placed in a URL.
func (h *WorkBuddyHandler) Issue(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	var req struct {
		KeyID int64  `json:"key_id" binding:"required,gt=0"`
		Model string `json:"model" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil || !workBuddyModelID.MatchString(req.Model) {
		response.BadRequest(c, "Choose an API key and a valid model ID")
		return
	}
	key, err := h.keys.GetByID(c.Request.Context(), req.KeyID)
	if err != nil || !eligibleWorkBuddyKey(key, subject.UserID) {
		response.BadRequest(c, "An active OpenAI group key owned by you is required")
		return
	}
	settings, err := h.settings.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.Error(c, 503, "Settings unavailable")
		return
	}
	endpoint, err := workBuddyEndpoint(settings.APIBaseURL)
	if err != nil {
		response.BadRequest(c, "Administrator must configure an HTTPS API Base URL first")
		return
	}
	token, err := h.store.Issue(c.Request.Context(), workbuddypair.Grant{
		UserID: subject.UserID, KeyID: key.ID, GroupID: key.Group.ID, Model: req.Model, URL: endpoint,
	})
	if err != nil {
		response.Error(c, 503, "Pairing unavailable; retry later")
		return
	}
	response.Success(c, gin.H{"code": token, "expires_in": int(workbuddypair.Lifetime.Seconds()), "endpoint": endpoint})
}

// Redeem consumes a bearer pairing code before releasing any credential. Re-read
// ownership/status so revocation or group reassignment after issuance takes effect.
func (h *WorkBuddyHandler) Redeem(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		response.BadRequest(c, "Invalid pairing request")
		return
	}
	grant, err := h.store.Consume(c.Request.Context(), req.Code)
	if err != nil {
		response.BadRequest(c, "Pairing code invalid, expired or already used")
		return
	}
	key, err := h.keys.GetByID(c.Request.Context(), grant.KeyID)
	if err != nil || !eligibleWorkBuddyKey(key, grant.UserID) || key.Group.ID != grant.GroupID {
		response.Forbidden(c, "API key is no longer available; generate a new pairing code")
		return
	}
	_, user, err := h.keys.ValidateKey(c.Request.Context(), key.Key)
	if err != nil || user == nil {
		response.Forbidden(c, "API key or user is no longer active")
		return
	}
	if h.settings != nil && h.settings.IsBackendModeEnabled(c.Request.Context()) && user.Role != "admin" {
		response.Forbidden(c, "User self-service is disabled")
		return
	}
	response.Success(c, gin.H{"schema_version": 1, "model": grant.Model, "url": grant.URL, "api_key": key.Key})
}
