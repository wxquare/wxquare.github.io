package domain

import "time"

type SupplySourceType string

const (
	SupplySourceLocalOps SupplySourceType = "LOCAL_OPS"
	SupplySourceMerchant SupplySourceType = "MERCHANT"
	SupplySourceSupplier SupplySourceType = "SUPPLIER"
)

type DraftStatus string

const (
	DraftStatusDraft     DraftStatus = "DRAFT"
	DraftStatusSubmitted DraftStatus = "SUBMITTED"
	DraftStatusDiscarded DraftStatus = "DISCARDED"
	DraftStatusArchived  DraftStatus = "ARCHIVED"
)

type StagingStatus string

const (
	StagingStatusValidated       StagingStatus = "VALIDATED"
	StagingStatusQCPending       StagingStatus = "QC_PENDING"
	StagingStatusQCReviewing     StagingStatus = "QC_REVIEWING"
	StagingStatusQCApproved      StagingStatus = "QC_APPROVED"
	StagingStatusApproved        StagingStatus = "APPROVED"
	StagingStatusPublishPending  StagingStatus = "PUBLISH_PENDING"
	StagingStatusPublished       StagingStatus = "PUBLISHED"
	StagingStatusRejected        StagingStatus = "REJECTED"
	StagingStatusWithdrawn       StagingStatus = "WITHDRAWN"
	StagingStatusCancelled       StagingStatus = "CANCELLED"
	StagingStatusVersionConflict StagingStatus = "VERSION_CONFLICT"
)

type QCReviewStatus string

const (
	QCReviewPending   QCReviewStatus = "PENDING"
	QCReviewReviewing QCReviewStatus = "REVIEWING"
	QCReviewApproved  QCReviewStatus = "APPROVED"
	QCReviewRejected  QCReviewStatus = "REJECTED"
	QCReviewCancelled QCReviewStatus = "CANCELLED"
	QCReviewPublished QCReviewStatus = "PUBLISHED"
)

type QCPolicy string

const (
	QCPolicyAutoApprove QCPolicy = "AUTO_APPROVE"
	QCPolicyRequired    QCPolicy = "QC_REQUIRED"
	QCPolicyBlock       QCPolicy = "BLOCK"
)

type ProductSupplyDraft struct {
	DraftID            string                `json:"draft_id"`
	OperationID        string                `json:"operation_id"`
	SourceType         SupplySourceType      `json:"source_type"`
	OperatorID         int64                 `json:"operator_id"`
	Status             DraftStatus           `json:"status"`
	BasePublishVersion int64                 `json:"base_publish_version"`
	Payload            ProductPublishPayload `json:"payload"`
	SourceStagingID    string                `json:"source_staging_id,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
}

type ProductSupplyStaging struct {
	StagingID          string                `json:"staging_id"`
	DraftID            string                `json:"draft_id"`
	OperationID        string                `json:"operation_id"`
	SourceType         SupplySourceType      `json:"source_type"`
	OperatorID         int64                 `json:"operator_id"`
	Status             StagingStatus         `json:"status"`
	QCPolicy           QCPolicy              `json:"qc_policy"`
	BasePublishVersion int64                 `json:"base_publish_version"`
	Payload            ProductPublishPayload `json:"payload"`
	ValidationErrors   []string              `json:"validation_errors,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
}

type ProductQCReview struct {
	ReviewID     string         `json:"review_id"`
	StagingID    string         `json:"staging_id"`
	Status       QCReviewStatus `json:"status"`
	ReviewPolicy QCPolicy       `json:"review_policy"`
	RiskReasons  []string       `json:"risk_reasons,omitempty"`
	ReviewerID   int64          `json:"reviewer_id,omitempty"`
	RejectReason string         `json:"reject_reason,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type SupplyOperationLog struct {
	LogID      string           `json:"log_id"`
	TraceID    string           `json:"trace_id"`
	ObjectType string           `json:"object_type"`
	ObjectID   string           `json:"object_id"`
	Action     string           `json:"action"`
	OperatorID int64            `json:"operator_id"`
	SourceType SupplySourceType `json:"source_type"`
	Message    string           `json:"message,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

type SupplyPublishResult struct {
	DraftID        string
	StagingID      string
	QCReviewID     string
	ItemID         int64
	SKUID          int64
	PublishVersion int64
	SnapshotID     string
	OutboxEventID  string
	InventoryKey   string
	InventoryReady bool
}
