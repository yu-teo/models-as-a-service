package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/logger"
	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/token"
)

// AccessLogger produces a structured JSON access log entry for each request.
// Includes tenant context when available (authenticated routes).
func AccessLogger() gin.HandlerFunc {
	log := logger.Production()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		tenantName := ""
		if u, ok := c.Get("user"); ok {
			if uc, ok := u.(*token.UserContext); ok {
				tenantName = uc.Tenant
			}
		}

		requestID := GetRequestID(c)

		log.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", requestID,
			"tenant_name", tenantName,
		)
	}
}
