package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/middleware"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func RegisterWorkBuddyRoutes(v1 *gin.RouterGroup, keys *service.APIKeyService, settings *service.SettingService,
	client *redis.Client, jwt servermiddleware.JWTAuthMiddleware, audit servermiddleware.AuditLogMiddleware) {
	h := handler.NewWorkBuddyHandler(keys, settings, client)
	limiter := middleware.NewRateLimiter(client)
	routes := v1.Group("/workbuddy")
	routes.POST("/pair", gin.HandlerFunc(jwt), servermiddleware.BackendModeUserGuard(settings), limiter.LimitWithOptions("workbuddy-pair", 10, time.Minute,
		middleware.RateLimitOptions{FailureMode: middleware.RateLimitFailClose}), gin.HandlerFunc(audit), h.Issue)
	routes.POST("/redeem", limiter.LimitWithOptions("workbuddy-redeem", 10, time.Minute,
		middleware.RateLimitOptions{FailureMode: middleware.RateLimitFailClose}), h.Redeem)
}
