package deleteurlsusecase

import (
	"time"
	
	"go-musthave-shortener/internal/config"
)

func DefaultConfig() config.DeleteURLsConfig {
	return config.DeleteURLsConfig{
		BufferSize:     256,
		BatchSize:      10,
		FlushInterval:  10 * time.Millisecond,
		WorkerCount:    2,
	}
}
