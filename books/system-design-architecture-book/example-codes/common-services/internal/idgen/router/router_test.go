package router

import (
	"context"
	"strings"
	"testing"
	"time"

	"common-services/internal/idgen"
	"common-services/internal/idgen/formatter"
	"common-services/internal/idgen/registry"
	"common-services/internal/idgen/segment"
	"common-services/internal/idgen/snowflake"
	ulidgen "common-services/internal/idgen/ulid"
)

func newTestService() *Service {
	reg := registry.NewStatic(registry.DefaultNamespaces())
	seg := segment.NewGenerator(segment.NewMemoryStore(map[string]int64{
		"product.sku": 600000000000,
	}, 100))
	clock := &snowflake.FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
	sf := snowflake.NewGenerator(snowflake.Config{Epoch: snowflake.DefaultEpoch()}, clock, &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1})
	return New(reg, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), 1000)
}

func TestNextBusinessOrder(t *testing.T) {
	svc := newTestService()
	got, err := svc.Next(context.Background(), idgen.IssueRequest{Namespace: "trade.order", Caller: "order-service", RequestID: "req-1", Now: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if !strings.HasPrefix(got.ID, "ORD20260429") {
		t.Fatalf("id = %s", got.ID)
	}
	if got.RawID == 0 {
		t.Fatal("raw id is zero")
	}
}

func TestBatchSegmentSKU(t *testing.T) {
	svc := newTestService()
	got, err := svc.Batch(context.Background(), idgen.IssueRequest{Namespace: "product.sku", Count: 3})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(got.IDs) != 3 {
		t.Fatalf("len = %d", len(got.IDs))
	}
	if got.IDs[0] != "600000000001" {
		t.Fatalf("first id = %s", got.IDs[0])
	}
}

func TestBatchTooLarge(t *testing.T) {
	svc := newTestService()
	_, err := svc.Batch(context.Background(), idgen.IssueRequest{Namespace: "product.sku", Count: 1001})
	svcErr, ok := idgen.AsServiceError(err)
	if !ok || svcErr.Code != idgen.ErrBatchTooLarge {
		t.Fatalf("err = %#v", err)
	}
}
