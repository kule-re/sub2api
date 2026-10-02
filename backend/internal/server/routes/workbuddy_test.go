package routes

import (
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkBuddyRouteGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1, DialTimeout: 100 * time.Millisecond, ReadTimeout: 100 * time.Millisecond, WriteTimeout: 100 * time.Millisecond})
	defer client.Close()
	router := gin.New()
	RegisterWorkBuddyRoutes(router.Group("/api/v1"), nil, nil, client,
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) { c.AbortWithStatus(401) }),
		servermiddleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }))
	request := func(path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/workbuddy/"+path, strings.NewReader(`{"code":"invalid"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w.Code
	}
	if request("pair") != 401 {
		t.Fatal("pair must use panel authentication")
	}
	for i := 0; i < 10; i++ {
		if request("redeem") != 400 {
			t.Fatal("unexpected redemption response")
		}
	}
	if request("redeem") != 429 {
		t.Fatal("public redemption must be rate limited")
	}
	r.Close()
	status := request("redeem")
	if status < 400 {
		t.Fatal("Redis failure must deny requests")
	}
}
