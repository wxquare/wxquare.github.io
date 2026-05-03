package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"common-services/internal/idgen/formatter"
	"common-services/internal/idgen/registry"
	"common-services/internal/idgen/router"
	"common-services/internal/idgen/segment"
	"common-services/internal/idgen/snowflake"
	ulidgen "common-services/internal/idgen/ulid"
	"common-services/internal/infrastructure/metrics"
)

func newHandlerForTest() http.Handler {
	reg := registry.NewStatic(registry.DefaultNamespaces())
	seg := segment.NewGenerator(segment.NewMemoryStore(map[string]int64{"product.sku": 600000000000}, 100))
	clock := &snowflake.FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
	sf := snowflake.NewGenerator(snowflake.Config{Epoch: snowflake.DefaultEpoch()}, clock, &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1})
	svc := router.New(reg, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), 1000)
	return NewHandler(svc, reg, metrics.NewRecorder(), func() bool { return true })
}

func TestNextID(t *testing.T) {
	h := newHandlerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ids/next", bytes.NewBufferString(`{"namespace":"trade.order","caller":"order-service","request_id":"req-1"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "ORD") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestUnknownNamespaceReturnsStructuredError(t *testing.T) {
	h := newHandlerForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ids/next", bytes.NewBufferString(`{"namespace":"missing","caller":"test","request_id":"req-2"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "NAMESPACE_NOT_FOUND") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestMetricsEndpoint(t *testing.T) {
	h := newHandlerForTest()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
