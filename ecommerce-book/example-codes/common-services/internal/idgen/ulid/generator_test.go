package ulid

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestGeneratorAddsPrefixAndULIDLength(t *testing.T) {
	g := NewGenerator()
	id, err := g.Next(context.Background(), "draft")
	if err != nil {
		t.Fatalf("next ulid: %v", err)
	}
	if !strings.HasPrefix(id, "draft_") {
		t.Fatalf("id = %s", id)
	}
	if len(id) != len("draft_")+26 {
		t.Fatalf("length = %d", len(id))
	}
}

func TestGeneratorConcurrentUniqueness(t *testing.T) {
	g := NewGenerator()
	const n = 1000
	seen := sync.Map{}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := g.Next(context.Background(), "evt")
			if err != nil {
				t.Errorf("next ulid: %v", err)
				return
			}
			if _, loaded := seen.LoadOrStore(id, true); loaded {
				t.Errorf("duplicate id: %s", id)
			}
		}()
	}
	wg.Wait()
}
