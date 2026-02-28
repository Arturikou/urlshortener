package models

type URLData struct {
	Alias       string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLRecord struct {
	ID          int64  `db:"id"`
	Alias       string `db:"alias"`
	URL         string `db:"url"`
	DeletedFlag bool   `db:"is_deleted"`
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
