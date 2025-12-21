package rabbitmq

import (
	"context"
	"math"
	"time"
)

// calculateBackoff calculates the backoff duration for a retry attempt.
// Uses exponential backoff: initialBackoff * (multiplier ^ attempt)
// Capped at maxBackoff.
func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	if attempt <= 0 {
		return cfg.InitialBackoff
	}

	backoff := float64(cfg.InitialBackoff) * math.Pow(cfg.Multiplier, float64(attempt))

	if backoff > float64(cfg.MaxBackoff) {
		return cfg.MaxBackoff
	}

	return time.Duration(backoff)
}

// shouldRetry determines if a message should be retried.
func shouldRetry(msg *Message, cfg RetryConfig) bool {
	return msg.RetryCount < cfg.MaxRetries
}

// retryHandler wraps a handler with retry logic.
type retryHandler struct {
	client  *Client
	handler HandlerFunc
	config  RetryConfig
	queue   string
}

// handle processes a message with retry logic.
func (r *retryHandler) handle(ctx context.Context, msg *Message) error {
	err := r.handler(ctx, msg)
	if err == nil {
		return nil
	}

	// Check if we should retry
	if !shouldRetry(msg, r.config) {
		// Max retries exceeded, send to DLQ
		return r.client.sendToDLQ(ctx, msg, r.queue, err)
	}

	// Calculate backoff and schedule retry
	backoff := calculateBackoff(msg.RetryCount, r.config)

	// Increment retry count
	retryMsg := msg.Clone()
	retryMsg.RetryCount++
	retryMsg.SetHeader("x-retry-reason", err.Error())
	retryMsg.SetHeader("x-retry-at", time.Now().Add(backoff).Format(time.RFC3339))

	// Wait for backoff duration
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(backoff):
	}

	// Re-publish for retry
	return r.client.PublishMessage(ctx, retryMsg)
}

// RetryInfo contains information about retry state.
type RetryInfo struct {
	Attempt    int           `json:"attempt"`
	MaxRetries int           `json:"max_retries"`
	NextBackoff time.Duration `json:"next_backoff"`
	Exhausted  bool          `json:"exhausted"`
}

// GetRetryInfo returns retry information for a message.
func GetRetryInfo(msg *Message, cfg RetryConfig) RetryInfo {
	exhausted := msg.RetryCount >= cfg.MaxRetries
	var nextBackoff time.Duration
	if !exhausted {
		nextBackoff = calculateBackoff(msg.RetryCount, cfg)
	}

	return RetryInfo{
		Attempt:     msg.RetryCount,
		MaxRetries:  cfg.MaxRetries,
		NextBackoff: nextBackoff,
		Exhausted:   exhausted,
	}
}
