package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

type HTTPReceiver struct {
	url    string
	client *http.Client
	mu     sync.Mutex
}

func NewHTTPReceiver(url string) (*HTTPReceiver, error) {
	if url == "" {
		return nil, nil
	}

	return &HTTPReceiver{
		url:    url,
		client: &http.Client{},
	}, nil
}

func (r *HTTPReceiver) Receive(event Event) error {
	if r == nil {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	resp, err := r.client.Post(r.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
