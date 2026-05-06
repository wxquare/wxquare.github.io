package audit

import (
	"context"
	"log"
	"time"
)

type IssueLog struct {
	RequestID    string
	Namespace    string
	Caller       string
	IssueType    string
	IssuedValue  string
	ErrorCode    string
	ErrorMessage string
	CreatedAt    time.Time
}

type Store interface {
	SaveIssueLog(ctx context.Context, entry IssueLog) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Record(ctx context.Context, entry IssueLog) {
	if s == nil || s.store == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if err := s.store.SaveIssueLog(ctx, entry); err != nil {
		log.Printf("[id-audit] save issue log failed: %v", err)
	}
}
