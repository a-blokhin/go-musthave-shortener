package createshortlinkbatchpkg

import "net/http"

const MethodPath = "/api/shorten/batch"
const Method = http.MethodPost

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchRequest []BatchRequestItem

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BatchResponse []BatchResponseItem