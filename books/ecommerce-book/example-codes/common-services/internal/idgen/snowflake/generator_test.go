package snowflake

import (
	"context"
	"testing"
	"time"

	"common-services/internal/idgen"
)

func TestGeneratorProducesIncreasingIDs(t *testing.T) {
	clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
	lease := &StaticLease{ReadyValue: true, WorkerIDValue: 3, RegionIDValue: 1}
	g := NewGenerator(Config{Epoch: DefaultEpoch(), MaxWaitRollback: time.Millisecond}, clock, lease)
	first, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second <= first {
		t.Fatalf("ids not increasing: %d <= %d", second, first)
	}
}

func TestGeneratorRejectsLeaseLost(t *testing.T) {
	clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
	lease := &StaticLease{ReadyValue: false}
	g := NewGenerator(Config{Epoch: DefaultEpoch()}, clock, lease)
	_, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
	svcErr, ok := idgen.AsServiceError(err)
	if !ok || svcErr.Code != idgen.ErrWorkerLeaseLost {
		t.Fatalf("err = %#v", err)
	}
}

func TestGeneratorRejectsLargeClockRollback(t *testing.T) {
	clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
	lease := &StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1}
	g := NewGenerator(Config{Epoch: DefaultEpoch(), MaxWaitRollback: time.Millisecond}, clock, lease)
	if _, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"}); err != nil {
		t.Fatalf("first: %v", err)
	}
	clock.NowValue = clock.NowValue.Add(-10 * time.Millisecond)
	_, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
	svcErr, ok := idgen.AsServiceError(err)
	if !ok || svcErr.Code != idgen.ErrClockRollback {
		t.Fatalf("err = %#v", err)
	}
}
