package models

type UserUrls struct {
	OriginalURL string `db:"url"`
	Alias       string `db:"alias"`
}
