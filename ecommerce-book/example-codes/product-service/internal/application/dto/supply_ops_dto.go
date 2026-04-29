package dto

type CreateSupplyDraftRequest struct {
	OperationID        string                   `json:"operation_id"`
	SourceType         string                   `json:"source_type"`
	OperatorID         int64                    `json:"operator_id"`
	BasePublishVersion int64                    `json:"base_publish_version"`
	Payload            ProductPublishPayloadDTO `json:"payload"`
}

type CreateSupplyDraftResponse struct {
	DraftID     string `json:"draft_id"`
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type SubmitSupplyDraftResponse struct {
	DraftID    string `json:"draft_id"`
	StagingID  string `json:"staging_id"`
	QCReviewID string `json:"qc_review_id,omitempty"`
	QCPolicy   string `json:"qc_policy"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type ApproveQCRequest struct {
	ReviewerID int64  `json:"reviewer_id"`
	Comment    string `json:"comment"`
}

type ApproveQCResponse struct {
	ReviewID  string `json:"review_id"`
	StagingID string `json:"staging_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type SupplyPublishResponse struct {
	StagingID      string `json:"staging_id"`
	ItemID         int64  `json:"item_id"`
	SKUID          int64  `json:"sku_id"`
	PublishVersion int64  `json:"publish_version"`
	SnapshotID     string `json:"snapshot_id"`
	OutboxEventID  string `json:"outbox_event_id"`
	InventoryKey   string `json:"inventory_key"`
	InventoryReady bool   `json:"inventory_ready"`
	Message        string `json:"message"`
}

type SupplyOperationLogDTO struct {
	LogID      string `json:"log_id"`
	TraceID    string `json:"trace_id"`
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	Action     string `json:"action"`
	OperatorID int64  `json:"operator_id"`
	SourceType string `json:"source_type"`
	Message    string `json:"message,omitempty"`
	CreatedAt  int64  `json:"created_at"`
}

type SupplyLogListResponse struct {
	Logs []SupplyOperationLogDTO `json:"logs"`
}
