package model

type ViewingProfile struct {
	UserID uint64            `json:"user_id"`
	Series []WatchSeries     `json:"series"`
	Genres []GenreWatchCount `json:"genres"`
}
