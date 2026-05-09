package segment

import (
	"context"
	"sync"
	"testing"

	"common-services/internal/idgen"
)

func TestGeneratorAllocatesSequentialIDs(t *testing.T) {
	store := NewMemoryStore(map[string]int64{"product.sku": 600000000000}, 3)
	g := NewGenerator(store)
	cfg := idgen.NamespaceConfig{Namespace: "product.sku", Step: 3}
	got := []int64{}
	for i := 0; i < 5; i++ {
		id, err := g.Next(context.Background(), cfg)
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		got = append(got, id)
	}
	want := []int64{600000000001, 600000000002, 600000000003, 600000000004, 600000000005}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%d want %d", i, got[i], want[i])
		}
	}
}

func TestGeneratorConcurrentUnique(t *testing.T) {
	store := NewMemoryStore(map[string]int64{"product.item": 800000000000}, 50)
	g := NewGenerator(store)
	cfg := idgen.NamespaceConfig{Namespace: "product.item", Step: 50}
	const n = 1000
	seen := sync.Map{}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := g.Next(context.Background(), cfg)
			if err != nil {
				t.Errorf("next: %v", err)
				return
			}
			if _, loaded := seen.LoadOrStore(id, true); loaded {
				t.Errorf("duplicate id %d", id)
			}
		}()
	}
	wg.Wait()
}
