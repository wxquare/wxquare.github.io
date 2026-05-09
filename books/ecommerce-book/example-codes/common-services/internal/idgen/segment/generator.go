package segment

import (
	"context"
	"sync"

	"common-services/internal/idgen"
)

type Range struct {
	Start int64
	End   int64
}

type Store interface {
	Allocate(ctx context.Context, namespace string, step int64) (Range, error)
}

type Generator struct {
	mu       sync.Mutex
	store    Store
	segments map[string]*state
}

type state struct {
	current int64
	end     int64
}

func NewGenerator(store Store) *Generator {
	return &Generator{store: store, segments: make(map[string]*state)}
}

func (g *Generator) Next(ctx context.Context, cfg idgen.NamespaceConfig) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	st := g.segments[cfg.Namespace]
	if st == nil || st.current >= st.end {
		step := cfg.Step
		if step <= 0 {
			step = 1000
		}
		allocated, err := g.store.Allocate(ctx, cfg.Namespace, step)
		if err != nil {
			return 0, idgen.NewError(idgen.ErrSegmentExhausted, cfg.Namespace, err.Error(), true)
		}
		st = &state{current: allocated.Start - 1, end: allocated.End}
		g.segments[cfg.Namespace] = st
	}

	st.current++
	return st.current, nil
}

type MemoryStore struct {
	mu    sync.Mutex
	maxID map[string]int64
	step  int64
}

func NewMemoryStore(initial map[string]int64, step int64) *MemoryStore {
	cp := make(map[string]int64, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &MemoryStore{maxID: cp, step: step}
}

func (s *MemoryStore) Allocate(ctx context.Context, namespace string, step int64) (Range, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if step <= 0 {
		step = s.step
	}
	start := s.maxID[namespace] + 1
	end := s.maxID[namespace] + step
	s.maxID[namespace] = end
	return Range{Start: start, End: end}, nil
}
