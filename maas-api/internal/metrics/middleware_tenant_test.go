package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/metrics"
	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/token"
)

// TestMiddleware_ExtractsTenantFromContext verifies that the metrics middleware
// reads the tenant from UserContext set by upstream auth middleware.
func TestMiddleware_ExtractsTenantFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := prometheus.NewRegistry()
	recorder, err := metrics.NewPrometheusRecorder(reg)
	require.NoError(t, err)

	router := gin.New()
	router.Use(metrics.NewMiddleware(recorder))
	router.GET("/v1/models", func(c *gin.Context) {
		c.Set("user", &token.UserContext{
			Username: "alice",
			Groups:   []string{"users"},
			Tenant:   "redteam",
		})
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	val := gatherMetricValue(t, reg, "maas_api_http_requests_total",
		map[string]string{"method": "GET", "route": "/v1/models", "status": "200", "tenant_name": "redteam"})
	assert.InDelta(t, float64(1), val, 0)
}

// TestMiddleware_EmptyTenantWhenNoUserContext verifies that requests without
// auth middleware (health, internal) produce metrics with an empty tenant label.
func TestMiddleware_EmptyTenantWhenNoUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := prometheus.NewRegistry()
	recorder, err := metrics.NewPrometheusRecorder(reg)
	require.NoError(t, err)

	router := gin.New()
	router.Use(metrics.NewMiddleware(recorder))
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	val := gatherMetricValue(t, reg, "maas_api_http_requests_total",
		map[string]string{"method": "GET", "route": "/health", "status": "200", "tenant_name": ""})
	assert.InDelta(t, float64(1), val, 0)
}
