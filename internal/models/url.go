package models

type URLData struct {
	Alias       string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
