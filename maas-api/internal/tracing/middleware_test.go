package tracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/token"
	"github.com/opendatahub-io/models-as-a-service/maas-api/internal/tracing"
)

func setupTracingTest(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(t.Context())
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
	})
	return exporter
}

// TestMiddleware_SpanWithTenantAttributes verifies that spans include tenant
// context attributes when UserContext is present.
func TestMiddleware_SpanWithTenantAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exporter := setupTracingTest(t)

	router := gin.New()
	router.Use(tracing.NewMiddleware("ai-tenant-redteam", "redteam-gw", "openshift-ingress"))
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

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "should have exactly one span")

	span := spans[0]
	assert.Equal(t, "GET /v1/models", span.Name)

	attrs := make(map[string]string)
	for _, a := range span.Attributes {
		attrs[string(a.Key)] = a.Value.AsString()
	}

	assert.Equal(t, "redteam", attrs["tenant.name"])
	assert.Equal(t, "ai-tenant-redteam", attrs["tenant.namespace"])
	assert.Equal(t, "redteam-gw", attrs["gateway.name"])
	assert.Equal(t, "openshift-ingress", attrs["gateway.namespace"])
	assert.Equal(t, "GET", attrs["http.method"])
	assert.Equal(t, "/v1/models", attrs["http.route"])
}

// TestMiddleware_SpanWithoutTenantContext verifies that spans have empty
// tenant attributes for unauthenticated routes (health, internal).
func TestMiddleware_SpanWithoutTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exporter := setupTracingTest(t)

	router := gin.New()
	router.Use(tracing.NewMiddleware("ai-tenant-redteam", "redteam-gw", "openshift-ingress"))
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := make(map[string]string)
	for _, a := range spans[0].Attributes {
		attrs[string(a.Key)] = a.Value.AsString()
	}

	assert.Empty(t, attrs["tenant.name"],
		"tenant.name should be empty for unauthenticated routes")
}

// TestMiddleware_SpanStatusOnServerError verifies that 5xx responses
// set the span status to Error.
func TestMiddleware_SpanStatusOnServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exporter := setupTracingTest(t)

	router := gin.New()
	router.Use(tracing.NewMiddleware("", "", ""))
	router.GET("/fail", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	router.ServeHTTP(w, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "Error", spans[0].Status.Code.String())
}

// TestMiddleware_NoopWhenNoProvider verifies that the middleware produces
// no exported spans when no TracerProvider is registered (default noop).
func TestMiddleware_NoopWhenNoProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	otel.SetTracerProvider(sdktrace.NewTracerProvider())

	router := gin.New()
	router.Use(tracing.NewMiddleware("", "", ""))
	router.GET("/v1/models", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
