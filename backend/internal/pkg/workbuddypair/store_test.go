package workbuddypair

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPairingLifecycle(t *testing.T) {
	r := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: r.Addr()})
	t.Cleanup(func() { c.Close() })
	s := Store{Client: c}
	ctx := context.Background()
	g := Grant{UserID: 4, KeyID: 9, GroupID: 2, Model: "example", URL: "https://example.com/v1/chat/completions"}
	token, err := s.Issue(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range r.Keys() {
		if strings.Contains(k, token) {
			t.Fatal("raw token persisted")
		}
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := s.Consume(ctx, token)
			if err == nil {
				successes.Add(1)
				if *got != g {
					t.Error("grant changed")
				}
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("redeemed %d times", successes.Load())
	}
	token, err = s.Issue(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	r.FastForward(Lifetime + time.Second)
	if _, err = s.Consume(ctx, token); err != ErrInvalid {
		t.Fatal("expired token accepted")
	}
	if _, err = s.Consume(ctx, "../malformed"); err != ErrInvalid {
		t.Fatal("malformed token accepted")
	}
	r.Close()
	if _, err = s.Issue(ctx, g); err == nil {
		t.Fatal("Redis failure must fail closed")
	}
}
