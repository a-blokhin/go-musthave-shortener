package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileReceiver struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

func NewFileReceiver(filePath string) (*FileReceiver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileReceiver{
		filePath: filePath,
		file:     file,
	}, nil
}

func (r *FileReceiver) Receive(event Event) error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = r.file.Write(append(data, '\n'))
	return err
}

func (r *FileReceiver) Close() error {
	if r == nil || r.file == nil {
		return nil
	}
	return r.file.Close()
}
