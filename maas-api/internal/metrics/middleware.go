package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/token"
)

func NewMiddleware(recorder MetricsRecorder, defaultTenant string) gin.HandlerFunc {
	if recorder == nil {
		panic("metrics.NewMiddleware: nil MetricsRecorder")
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		recorder.IncrementInFlight(method)
		defer recorder.DecrementInFlight(method)
		start := time.Now()

		c.Next()
		duration := time.Since(start)

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		tenant := defaultTenant
		if u, ok := c.Get("user"); ok {
			if uc, ok := u.(*token.UserContext); ok {
				tenant = uc.Tenant
			}
		}

		status := strconv.Itoa(c.Writer.Status())
		recorder.RecordRequestDuration(method, route, status, tenant, duration)
	}
}
