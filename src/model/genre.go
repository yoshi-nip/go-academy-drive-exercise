package model

type GenreWatchCount struct {
	GenreID     string `json:"genre_id"`
	SeriesCount uint64 `json:"series_count"`
}
