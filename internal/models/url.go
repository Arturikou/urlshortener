package models

type URLData struct {
	Alias       string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UpsertResult struct {
	Alias    string
	IsInsert bool
}

type ShortenBatch struct {
	CorrelationID string
	OriginalURL   string
	Alias         string
}
