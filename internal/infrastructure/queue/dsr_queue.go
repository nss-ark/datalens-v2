package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// DSRQueue defines the interface for queuing DSR execution jobs.
type DSRQueue interface {
	// Enqueue publishes a DSR ID to the queue for execution.
	Enqueue(ctx context.Context, dsrID string) error

	// Subscribe registers a handler to process DSR execution jobs.
	Subscribe(ctx context.Context, handler func(ctx context.Context, dsrID string) error) error
}

// NATSDSRQueue implements DSRQueue using NATS JetStream.
type NATSDSRQueue struct {
	js     jetstream.JetStream
	logger *slog.Logger
	stream string
}

const (
	DSRStreamName    = "DATALENS_DSR_EXECUTION"
	DSRStreamSubject = "dsr.execution.>"
	DSRJobSubject    = "dsr.execution.requested"
	DSRConsumerName  = "dsr-executor-worker"
)

// NewNATSDSRQueue creates a new NATSDSRQueue.
func NewNATSDSRQueue(conn *nats.Conn, logger *slog.Logger) (*NATSDSRQueue, error) {
	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("init jetstream: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure Stream exists
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      DSRStreamName,
		Subjects:  []string{DSRStreamSubject},
		Retention: jetstream.WorkQueuePolicy, // Remove msg after ack
		Storage:   jetstream.FileStorage,
	})
	if err != nil {
		return nil, fmt.Errorf("create dsr stream: %w", err)
	}

	return &NATSDSRQueue{
		js:     js,
		logger: logger.With("component", "dsr_queue"),
		stream: DSRStreamName,
	}, nil
}

// Enqueue publishes the DSR ID to the NATS subject.
func (q *NATSDSRQueue) Enqueue(ctx context.Context, dsrID string) error {
	payload, err := json.Marshal(dsrID)
	if err != nil {
		return fmt.Errorf("marshal dsr id: %w", err)
	}

	_, err = q.js.Publish(ctx, DSRJobSubject, payload)
	if err != nil {
		return fmt.Errorf("publish dsr job: %w", err)
	}

	q.logger.Debug("dsr execution job enqueued", "dsr_id", dsrID)
	return nil
}

// Subscribe starts a consumer to process DSR execution jobs.
func (q *NATSDSRQueue) Subscribe(ctx context.Context, handler func(ctx context.Context, dsrID string) error) error {
	// Create durable consumer
	consumer, err := q.js.CreateOrUpdateConsumer(ctx, DSRStreamName, jetstream.ConsumerConfig{
		Durable:       DSRConsumerName,
		FilterSubject: DSRJobSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("create dsr consumer: %w", err)
	}

	// Consume messages
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		var dsrID string
		if err := json.Unmarshal(msg.Data(), &dsrID); err != nil {
			q.logger.Error("invalid dsr job payload", "error", err)
			msg.Term()
			return
		}

		q.logger.Info("processing dsr execution", "dsr_id", dsrID)

		// Create context for handler
		hCtx := context.Background()

		if err := handler(hCtx, dsrID); err != nil {
			q.logger.Error("dsr execution failed", "dsr_id", dsrID, "error", err)
			msg.Nak()
			return
		}

		msg.Ack()
		q.logger.Info("dsr execution completed", "dsr_id", dsrID)
	})

	if err != nil {
		return fmt.Errorf("consume dsr jobs: %w", err)
	}

	_ = cons // Keep reference to prevent GC
	return nil
}
