package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/logger"
	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/token"
)

const loggerKey = "logger"

// TenantLogger returns a Gin middleware that creates a per-request logger
// enriched with tenant context and request ID. The enriched logger is stored
// in the Gin context and accessible via GetLogger.
func TenantLogger(base *logger.Logger) gin.HandlerFunc {
	if base == nil {
		base = logger.Production()
	}
	return func(c *gin.Context) {
		fields := []any{}

		if requestID := GetRequestID(c); requestID != "" {
			fields = append(fields, "request_id", requestID)
		}

		if u, ok := c.Get("user"); ok {
			if uc, ok := u.(*token.UserContext); ok {
				fields = append(fields, "tenant_name", uc.Tenant)
			}
		}

		enriched := base.WithFields(fields...)
		c.Set(loggerKey, enriched)
		c.Next()
	}
}

// GetLogger retrieves the per-request logger from the Gin context.
// Falls back to a production logger if none is set.
func GetLogger(c *gin.Context) *logger.Logger {
	if l, ok := c.Get(loggerKey); ok {
		if log, ok := l.(*logger.Logger); ok {
			return log
		}
	}
	return logger.Production()
}
