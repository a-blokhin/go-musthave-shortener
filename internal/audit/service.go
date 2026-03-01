package audit

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	receivers []Receiver
	logger    *zap.Logger
	mu        sync.RWMutex
}

func NewService(logger *zap.Logger) *Service {
	return &Service{
		receivers: make([]Receiver, 0),
		logger:    logger,
	}
}

func (s *Service) AddReceiver(receiver Receiver) {
	if receiver == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receivers = append(s.receivers, receiver)
}

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
