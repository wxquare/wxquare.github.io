package persistence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"product-service/internal/domain"
)

type SupplyOpsRepository struct {
	mu          sync.RWMutex
	seq         int64
	drafts      map[string]*domain.ProductSupplyDraft
	stagings    map[string]*domain.ProductSupplyStaging
	qcReviews   map[string]*domain.ProductQCReview
	qcByStaging map[string]string
	logs        []domain.SupplyOperationLog
}

func NewSupplyOpsRepository() *SupplyOpsRepository {
	return &SupplyOpsRepository{
		drafts:      make(map[string]*domain.ProductSupplyDraft),
		stagings:    make(map[string]*domain.ProductSupplyStaging),
		qcReviews:   make(map[string]*domain.ProductQCReview),
		qcByStaging: make(map[string]string),
		logs:        make([]domain.SupplyOperationLog, 0),
	}
}

func (r *SupplyOpsRepository) NextID(ctx context.Context, prefix string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), r.seq)
}

func (r *SupplyOpsRepository) SaveDraft(ctx context.Context, draft *domain.ProductSupplyDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drafts[draft.DraftID] = cloneDraft(draft)
	return nil
}

func (r *SupplyOpsRepository) GetDraft(ctx context.Context, draftID string) (*domain.ProductSupplyDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	draft, ok := r.drafts[draftID]
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", draftID)
	}
	return cloneDraft(draft), nil
}

func (r *SupplyOpsRepository) SaveStaging(ctx context.Context, staging *domain.ProductSupplyStaging) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stagings[staging.StagingID] = cloneStaging(staging)
	return nil
}

func (r *SupplyOpsRepository) GetStaging(ctx context.Context, stagingID string) (*domain.ProductSupplyStaging, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	staging, ok := r.stagings[stagingID]
	if !ok {
		return nil, fmt.Errorf("staging not found: %s", stagingID)
	}
	return cloneStaging(staging), nil
}

func (r *SupplyOpsRepository) SaveQCReview(ctx context.Context, review *domain.ProductQCReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.qcReviews[review.ReviewID] = cloneQCReview(review)
	r.qcByStaging[review.StagingID] = review.ReviewID
	return nil
}

func (r *SupplyOpsRepository) GetQCReview(ctx context.Context, reviewID string) (*domain.ProductQCReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	review, ok := r.qcReviews[reviewID]
	if !ok {
		return nil, fmt.Errorf("qc review not found: %s", reviewID)
	}
	return cloneQCReview(review), nil
}

func (r *SupplyOpsRepository) GetQCReviewByStaging(ctx context.Context, stagingID string) (*domain.ProductQCReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reviewID, ok := r.qcByStaging[stagingID]
	if !ok {
		return nil, fmt.Errorf("qc review not found for staging: %s", stagingID)
	}
	review, ok := r.qcReviews[reviewID]
	if !ok {
		return nil, fmt.Errorf("qc review not found: %s", reviewID)
	}
	return cloneQCReview(review), nil
}

func (r *SupplyOpsRepository) AppendLog(ctx context.Context, log domain.SupplyOperationLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, log)
	return nil
}

func (r *SupplyOpsRepository) ListLogs(ctx context.Context, traceID string) ([]domain.SupplyOperationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.SupplyOperationLog, 0)
	for _, log := range r.logs {
		if traceID == "" || log.TraceID == traceID {
			result = append(result, log)
		}
	}
	return result, nil
}

func cloneDraft(draft *domain.ProductSupplyDraft) *domain.ProductSupplyDraft {
	if draft == nil {
		return nil
	}
	cp := *draft
	cp.Payload = clonePublishPayload(draft.Payload)
	return &cp
}

func cloneStaging(staging *domain.ProductSupplyStaging) *domain.ProductSupplyStaging {
	if staging == nil {
		return nil
	}
	cp := *staging
	cp.Payload = clonePublishPayload(staging.Payload)
	cp.ValidationErrors = append([]string(nil), staging.ValidationErrors...)
	return &cp
}

func cloneQCReview(review *domain.ProductQCReview) *domain.ProductQCReview {
	if review == nil {
		return nil
	}
	cp := *review
	cp.RiskReasons = append([]string(nil), review.RiskReasons...)
	return &cp
}

func clonePublishPayload(payload domain.ProductPublishPayload) domain.ProductPublishPayload {
	cp := payload
	cp.Attributes = cloneStringMap(payload.Attributes)
	cp.Images = append([]string(nil), payload.Images...)
	cp.Resource.Attributes = cloneStringMap(payload.Resource.Attributes)
	cp.Offer.Channels = append([]string(nil), payload.Offer.Channels...)
	cp.Offer.Attributes = cloneStringMap(payload.Offer.Attributes)
	cp.InputSchema.Fields = append([]domain.InputField(nil), payload.InputSchema.Fields...)
	cp.Fulfillment.Attributes = cloneStringMap(payload.Fulfillment.Attributes)
	cp.RefundRule.Attributes = cloneStringMap(payload.RefundRule.Attributes)
	return cp
}
