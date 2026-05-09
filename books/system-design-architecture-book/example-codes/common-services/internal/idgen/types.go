package idgen

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type IDType string
type GeneratorType string
type ExposeScope string
type NamespaceStatus string
type ErrorCode string

const (
	IDTypeInt64      IDType = "INT64"
	IDTypeString     IDType = "STRING"
	IDTypeBusinessNo IDType = "BUSINESS_NO"

	GeneratorSegment   GeneratorType = "SEGMENT"
	GeneratorSnowflake GeneratorType = "SNOWFLAKE"
	GeneratorULID      GeneratorType = "ULID"

	ExposeInternal ExposeScope = "INTERNAL"
	ExposeExternal ExposeScope = "EXTERNAL"
	ExposeMixed    ExposeScope = "MIXED"

	NamespaceEnabled    NamespaceStatus = "ENABLED"
	NamespaceDisabled   NamespaceStatus = "DISABLED"
	NamespaceDeprecated NamespaceStatus = "DEPRECATED"

	ErrNamespaceNotFound ErrorCode = "NAMESPACE_NOT_FOUND"
	ErrNamespaceDisabled ErrorCode = "NAMESPACE_DISABLED"
	ErrGeneratorNotReady ErrorCode = "GENERATOR_NOT_READY"
	ErrSegmentExhausted  ErrorCode = "SEGMENT_EXHAUSTED"
	ErrWorkerLeaseLost   ErrorCode = "WORKER_LEASE_LOST"
	ErrClockRollback     ErrorCode = "CLOCK_ROLLBACK"
	ErrBatchTooLarge     ErrorCode = "BATCH_TOO_LARGE"
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
)

type NamespaceConfig struct {
	Namespace     string          `json:"namespace"`
	BizDomain     string          `json:"biz_domain"`
	IDType        IDType          `json:"id_type"`
	GeneratorType GeneratorType   `json:"generator_type"`
	Prefix        string          `json:"prefix"`
	ExposeScope   ExposeScope     `json:"expose_scope"`
	Step          int64           `json:"step"`
	MaxCapacity   int64           `json:"max_capacity"`
	OwnerTeam     string          `json:"owner_team"`
	Status        NamespaceStatus `json:"status"`
}

type IssueRequest struct {
	Namespace string
	Caller    string
	RequestID string
	Count     int
	Now       time.Time
}

type IssueResult struct {
	Namespace string        `json:"namespace"`
	IDType    IDType        `json:"id_type"`
	Generator GeneratorType `json:"generator"`
	RawID     int64         `json:"raw_id,omitempty"`
	ID        string        `json:"id"`
	IssuedAt  time.Time     `json:"issued_at"`
}

type BatchResult struct {
	Namespace string        `json:"namespace"`
	IDType    IDType        `json:"id_type"`
	Generator GeneratorType `json:"generator"`
	Count     int           `json:"count"`
	IDs       []string      `json:"ids"`
	RawIDs    []int64       `json:"raw_ids,omitempty"`
	IssuedAt  time.Time     `json:"issued_at"`
}

type ServiceError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Namespace string    `json:"namespace,omitempty"`
	Retryable bool      `json:"retryable"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, namespace, message string, retryable bool) *ServiceError {
	return &ServiceError{Code: code, Namespace: namespace, Message: message, Retryable: retryable}
}

func AsServiceError(err error) (*ServiceError, bool) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr, true
	}
	return nil, false
}

type Registry interface {
	Get(ctx context.Context, namespace string) (NamespaceConfig, error)
	List(ctx context.Context) ([]NamespaceConfig, error)
}

type Service interface {
	Next(ctx context.Context, req IssueRequest) (IssueResult, error)
	Batch(ctx context.Context, req IssueRequest) (BatchResult, error)
}
