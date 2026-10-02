// Package workbuddypair stores short-lived, single-use WorkBuddy setup grants.
// Neither API keys nor raw pairing tokens are persisted here.
package workbuddypair

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

const Lifetime = 5 * time.Minute

var ErrInvalid = errors.New("pairing code is invalid, expired or already used")
var tokenPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Grant struct {
	UserID  int64  `json:"user_id"`
	KeyID   int64  `json:"key_id"`
	GroupID int64  `json:"group_id"`
	Model   string `json:"model"`
	URL     string `json:"url"`
}

type Store struct{ Client *redis.Client }

func tokenKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "workbuddy:pair:" + hex.EncodeToString(sum[:])
}

func (s Store) Issue(ctx context.Context, grant Grant) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	data, err := json.Marshal(grant)
	if err != nil {
		return "", err
	}
	ok, err := s.Client.SetNX(ctx, tokenKey(token), data, Lifetime).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("pairing collision")
	}
	return token, nil
}

func (s Store) Consume(ctx context.Context, token string) (*Grant, error) {
	if !tokenPattern.MatchString(token) {
		return nil, ErrInvalid
	}
	// GETDEL is atomic: concurrent redeemers cannot receive the same grant.
	data, err := s.Client.GetDel(ctx, tokenKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	var grant Grant
	if err := json.Unmarshal(data, &grant); err != nil {
		return nil, err
	}
	return &grant, nil
}
