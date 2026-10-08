package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OutboxRecord represents a row in the outbox_events table.
type OutboxRecord struct {
	ID            int64      `db:"id" json:"id"`
	AggregateType string     `db:"aggregate_type" json:"aggregate_type"`
	AggregateID   string     `db:"aggregate_id" json:"aggregate_id"`
	EventID       string     `db:"event_id" json:"event_id"`
	EventType     string     `db:"event_type" json:"event_type"`
	EventVersion  int32      `db:"event_version" json:"event_version"`
	Payload       []byte     `db:"payload" json:"payload"`
	Status        string     `db:"status" json:"status"`
	Attempts      int32      `db:"attempts" json:"attempts"`
	LastError     *string    `db:"last_error" json:"last_error"`
	AvailableAt   time.Time  `db:"available_at" json:"available_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	PublishedAt   *time.Time `db:"published_at" json:"published_at"`
}

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

func normalizeOutboxRecord(record OutboxRecord) OutboxRecord {
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
	return record
}

// InsertOutboxDB inserts a single event into the outbox. The insert runs
// outside any transaction — it is the caller's responsibility to ensure the
// outbox row is inserted atomically within the same DB transaction as the
// business mutation.
func InsertOutboxDB(ctx context.Context, db *sqlx.DB, record OutboxRecord) error {
	return insertOutbox(ctx, db, record)
}

// InsertOutboxTx inserts a single event into the outbox within the given
// transaction, keeping the event atomic with the business mutation.
func InsertOutboxTx(ctx context.Context, tx *sqlx.Tx, record OutboxRecord) error {
	return insertOutbox(ctx, tx, record)
}

func insertOutbox(ctx context.Context, q sqlx.ExtContext, record OutboxRecord) error {
	record = normalizeOutboxRecord(record)

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
	`

	_, err := q.ExecContext(ctx, query,
		record.AggregateType,
		record.AggregateID,
		record.EventID,
		record.EventType,
		record.EventVersion,
		record.Payload,
		record.Status,
		record.AvailableAt,
	)
	return err
}

// ClaimPendingOutbox locks up to maxRows pending/retriable outbox rows
// and returns them for publishing. Uses FOR UPDATE SKIP LOCKED so
// multiple publisher instances can safely poll concurrently.
// Stale 'processing' rows (crashed publisher) are also reclaimed after
// a grace period.
func ClaimPendingOutbox(ctx context.Context, q sqlx.ExtContext, maxRows int32, maxAttempts int32) ([]*OutboxRecord, error) {
	const query = `
		UPDATE outbox_events
		SET status = 'processing',
			attempts = attempts + 1,
			available_at = now() + interval '5 minutes'
		WHERE id IN (
			SELECT id FROM outbox_events
			WHERE (
			    status IN ('pending', 'failed')
			    OR (status = 'processing' AND available_at <= now() - interval '10 minutes')
			)
			  AND attempts < $1
			  AND available_at <= now()
			ORDER BY available_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, aggregate_type, aggregate_id, event_id,
			event_type, event_version, payload, status, attempts,
			last_error, available_at, created_at, published_at
	`

	var records []*OutboxRecord
	err := sqlx.SelectContext(ctx, q, &records, query, maxAttempts, maxRows)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CompleteOutbox marks an outbox event as successfully published.
func CompleteOutbox(ctx context.Context, q sqlx.ExtContext, eventID string) error {
	const query = `
		UPDATE outbox_events
		SET status = $1, published_at = now()
		WHERE event_id = $2
	`
	_, err := q.ExecContext(ctx, query, OutboxStatusPublished, eventID)
	return err
}

// FailOutbox marks an outbox event as failed with error details.
func FailOutbox(ctx context.Context, q sqlx.ExtContext, eventID string, errMsg string) error {
	const query = `
		UPDATE outbox_events
		SET status = $1, last_error = $2, available_at = now() + interval '30 seconds'
		WHERE event_id = $3
	`
	_, err := q.ExecContext(ctx, query, OutboxStatusFailed, errMsg, eventID)
	return err
}

// OutboxEnvelope standardizes the event payload with metadata.
type OutboxEnvelope struct {
	EventID       string          `db:"-" json:"event_id"`
	EventType     string          `db:"-" json:"event_type"`
	EventVersion  int32           `db:"-" json:"event_version"`
	AggregateID   string          `db:"-" json:"aggregate_id"`
	OccurredAt    time.Time       `db:"-" json:"occurred_at"`
	CorrelationID string          `db:"-" json:"correlation_id,omitempty"`
	Payload       json.RawMessage `db:"-" json:"payload"`
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

// CountPendingOutbox returns the number of events still awaiting delivery
// (pending or retriable failed). Used to expose outbox lag as a gauge.
func CountPendingOutbox(ctx context.Context, q sqlx.ExtContext) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM outbox_events
		WHERE status IN ('pending', 'failed')
	`
	var count int64
	err := sqlx.GetContext(ctx, q, &count, query)
	return count, err
}

// ErrOutboxNoRows is returned when an update affects zero rows.
var ErrOutboxNoRows = sql.ErrNoRows

// ErrOutboxNotFound is a sentinel for missing outbox records.
var ErrOutboxNotFound = errors.New("outbox record not found")
