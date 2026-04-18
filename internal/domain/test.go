package domain

import (
	"github.com/google/uuid"
	"time"
)

type statusTest string

const (
	StatusRunning   statusTest = "running"
	StatusFinished  statusTest = "finished"
	StatusFailed    statusTest = "failed"
	StatusCancelled statusTest = "cancelled"
)

type Test struct {
	ID              uuid.UUID  `json:"id"`
	URL             string     `json:"url"`
	RPS             int        `json:"rps"`
	DurationSeconds int        `json:"duration_seconds"`
	Status          statusTest `json:"status"`
}

type Metric struct {
	ID            uuid.UUID `json:"id"`
	TestID        uuid.UUID `json:"test_id"`
	BucketTime    time.Time `json:"bucket_Time"`
	RequestsCount int       `json:"requests_count"`
	SuccessCount  int       `json:"success_count"`
	ErrorCount    int       `json:"error_count"`
	AVGLatencyMs  float64   `json:"avg_latency_ms"`
	P95LatencyMs  float64   `json:"p95_latency_ms"`
}
