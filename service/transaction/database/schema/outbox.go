package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OutboxRecord represents a row in the outbox_events table.
type OutboxRecord struct {
	ID            int64      `json:"id" db:"id"`
	AggregateType string     `json:"aggregate_type" db:"aggregate_type"`
	AggregateID   string     `json:"aggregate_id" db:"aggregate_id"`
	EventID       string     `json:"event_id" db:"event_id"`
	EventType     string     `json:"event_type" db:"event_type"`
	EventVersion  int32      `json:"event_version" db:"event_version"`
	Payload       []byte     `json:"payload" db:"payload"`
	Status        string     `json:"status" db:"status"`
	Attempts      int32      `json:"attempts" db:"attempts"`
	LastError     *string    `json:"last_error" db:"last_error"`
	AvailableAt   time.Time  `json:"available_at" db:"available_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	PublishedAt   *time.Time `json:"published_at" db:"published_at"`
}

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

// InsertOutboxDB inserts a single event into the outbox over a *sqlx.DB.
// The insert runs outside any transaction — it is the caller's
// responsibility to ensure the outbox row is inserted atomically
// within the same DB transaction as the business mutation.
func InsertOutboxDB(ctx context.Context, database *sqlx.DB, record OutboxRecord) error {
	return insertOutbox(ctx, database, record)
}

// InsertOutboxTx inserts a single event into the outbox within an *sqlx.Tx,
// keeping the outbox row atomic with the business mutation.
func InsertOutboxTx(ctx context.Context, tx *sqlx.Tx, record OutboxRecord) error {
	return insertOutbox(ctx, tx, record)
}

func insertOutbox(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, record OutboxRecord) error {
	if record.EventID == "" {
		record.EventID = uuid.New().String()
	}
	if record.EventVersion == 0 {
		record.EventVersion = 1
	}
	if record.Status == "" {
		record.Status = OutboxStatusPending
	}
	if record.AvailableAt.IsZero() {
		record.AvailableAt = time.Now()
	}

	const query = `
		INSERT INTO outbox_events (
			aggregate_type, aggregate_id, event_id, event_type,
			event_version, payload, status, attempts,
			last_error, available_at, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, 0,
			NULL, $8, now()
		)
		RETURNING id, created_at
	`

	return q.QueryRowContext(ctx, query,
		record.AggregateType,
		record.AggregateID,
		record.EventID,
		record.EventType,
		record.EventVersion,
		record.Payload,
		record.Status,
		record.AvailableAt,
	).Scan(&record.ID, &record.CreatedAt)
}

// OutboxEnvelope standardizes the event payload with metadata.
type OutboxEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventVersion  int32           `json:"event_version"`
	AggregateID   string          `json:"aggregate_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// MarshalEnvelope wraps a payload into the standard envelope.
func MarshalEnvelope(eventID, eventType string, version int32, aggregateID string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	envelope := OutboxEnvelope{
		EventID:      eventID,
		EventType:    eventType,
		EventVersion: version,
		AggregateID:  aggregateID,
		OccurredAt:   time.Now(),
		Payload:      payloadBytes,
	}

	return json.Marshal(envelope)
}
