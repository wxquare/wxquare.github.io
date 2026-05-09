package router

import (
	"context"
	"strconv"
	"time"

	"common-services/internal/idgen"
	"common-services/internal/idgen/formatter"
	"common-services/internal/idgen/segment"
	"common-services/internal/idgen/snowflake"
	ulidgen "common-services/internal/idgen/ulid"
)

type Service struct {
	registry  idgen.Registry
	segment   *segment.Generator
	snowflake *snowflake.Generator
	ulid      *ulidgen.Generator
	formatter *formatter.BusinessNumberFormatter
	maxBatch  int
}

func New(reg idgen.Registry, seg *segment.Generator, sf *snowflake.Generator, ug *ulidgen.Generator, fmt *formatter.BusinessNumberFormatter, maxBatch int) *Service {
	if maxBatch <= 0 {
		maxBatch = 1000
	}
	return &Service{registry: reg, segment: seg, snowflake: sf, ulid: ug, formatter: fmt, maxBatch: maxBatch}
}

func (s *Service) Next(ctx context.Context, req idgen.IssueRequest) (idgen.IssueResult, error) {
	cfg, err := s.registry.Get(ctx, req.Namespace)
	if err != nil {
		return idgen.IssueResult{}, err
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	raw, id, err := s.issueOne(ctx, cfg, now)
	if err != nil {
		return idgen.IssueResult{}, err
	}
	return idgen.IssueResult{Namespace: cfg.Namespace, IDType: cfg.IDType, Generator: cfg.GeneratorType, RawID: raw, ID: id, IssuedAt: now}, nil
}

func (s *Service) Batch(ctx context.Context, req idgen.IssueRequest) (idgen.BatchResult, error) {
	if req.Count <= 0 {
		return idgen.BatchResult{}, idgen.NewError(idgen.ErrInvalidRequest, req.Namespace, "count must be positive", false)
	}
	if req.Count > s.maxBatch {
		return idgen.BatchResult{}, idgen.NewError(idgen.ErrBatchTooLarge, req.Namespace, "count exceeds max batch size", false)
	}
	cfg, err := s.registry.Get(ctx, req.Namespace)
	if err != nil {
		return idgen.BatchResult{}, err
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	ids := make([]string, 0, req.Count)
	raws := make([]int64, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		raw, id, err := s.issueOne(ctx, cfg, now)
		if err != nil {
			return idgen.BatchResult{}, err
		}
		ids = append(ids, id)
		if raw != 0 {
			raws = append(raws, raw)
		}
	}
	return idgen.BatchResult{Namespace: cfg.Namespace, IDType: cfg.IDType, Generator: cfg.GeneratorType, Count: req.Count, IDs: ids, RawIDs: raws, IssuedAt: now}, nil
}

func (s *Service) issueOne(ctx context.Context, cfg idgen.NamespaceConfig, now time.Time) (int64, string, error) {
	switch cfg.GeneratorType {
	case idgen.GeneratorSegment:
		raw, err := s.segment.Next(ctx, cfg)
		if err != nil {
			return 0, "", err
		}
		return raw, strconv.FormatInt(raw, 10), nil
	case idgen.GeneratorSnowflake:
		raw, err := s.snowflake.Next(ctx, cfg)
		if err != nil {
			return 0, "", err
		}
		if cfg.IDType == idgen.IDTypeBusinessNo {
			return raw, s.formatter.Format(cfg.Prefix, raw, now), nil
		}
		return raw, strconv.FormatInt(raw, 10), nil
	case idgen.GeneratorULID:
		id, err := s.ulid.Next(ctx, cfg.Prefix)
		return 0, id, err
	default:
		return 0, "", idgen.NewError(idgen.ErrInvalidRequest, cfg.Namespace, "unsupported generator type", false)
	}
}
