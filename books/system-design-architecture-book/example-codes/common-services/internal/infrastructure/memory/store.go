package memory

import (
	"context"
	"sync"

	"common-services/internal/idgen"
	"common-services/internal/idgen/registry"
	"common-services/internal/idgen/segment"
	"common-services/internal/infrastructure/audit"
)

type Store struct {
	*registry.StaticRegistry
	Segment *segment.MemoryStore
	mu      sync.Mutex
	Logs    []audit.IssueLog
}

func NewStore() *Store {
	initial := map[string]int64{
		"product.item": 800000000000,
		"product.spu":  700000000000,
		"product.sku":  600000000000,
	}
	return &Store{
		StaticRegistry: registry.NewStatic(registry.DefaultNamespaces()),
		Segment:        segment.NewMemoryStore(initial, 1000),
		Logs:           make([]audit.IssueLog, 0),
	}
}

func (s *Store) SaveIssueLog(ctx context.Context, entry audit.IssueLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Logs = append(s.Logs, entry)
	return nil
}

var _ idgen.Registry = (*Store)(nil)
