// Package audit provides audit logging functionality for the URL shortener service.
package audit

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Service manages audit event receivers and emits events to all registered receivers.
type Service struct {
	receivers []Receiver
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewService creates a new audit service with the provided logger.
//
// Parameters:
//   - logger: zap logger for logging audit service events
//
// Returns a new Service instance ready to receive and emit audit events.
func NewService(logger *zap.Logger) *Service {
	return &Service{
		receivers: make([]Receiver, 0),
		logger:    logger,
	}
}

// AddReceiver adds a new receiver to the audit service.
// The receiver will receive all future audit events.
//
// Parameters:
//   - receiver: the receiver to add (nil receivers are ignored)
func (s *Service) AddReceiver(receiver Receiver) {
	if receiver == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receivers = append(s.receivers, receiver)
}

// Emit creates and sends an audit event to all registered receivers.
// The event is sent asynchronously to each receiver in a separate goroutine.
// This method is thread-safe and non-blocking.
//
// Parameters:
//   - action: the action being audited (e.g., "create", "delete", "redirect")
//   - userID: the user ID performing the action
//   - url: the URL being acted upon
func (s *Service) Emit(action, userID, url string) {
	event := Event{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}

	s.mu.RLock()
	receivers := make([]Receiver, len(s.receivers))
	copy(receivers, s.receivers)
	s.mu.RUnlock()

	for _, receiver := range receivers {
		go func(r Receiver) {
			if err := r.Receive(event); err != nil {
				s.logger.Error("Failed to send audit event",
					zap.Error(err),
					zap.String("action", action),
					zap.String("url", url))
			}
		}(receiver)
	}
}

// Close closes all receivers that implement the Close() method.
// This should be called when shutting down the service to ensure
// all audit events are flushed and resources are released.
func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, receiver := range s.receivers {
		if closer, ok := receiver.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				s.logger.Error("Failed to close audit receiver", zap.Error(err))
			}
		}
	}
	s.receivers = nil
}
