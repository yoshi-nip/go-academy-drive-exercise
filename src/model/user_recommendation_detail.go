package model

type UserRecommendationDetail struct {
	UserID            uint64        `json:"user_id"`
	OtherUserID       uint64        `json:"other_user_id"`
	CommonSeries      []WatchSeries `json:"common_series"`
	CommonGenres      []Genre       `json:"common_genres"`
	RecommendedSeries []WatchSeries `json:"recommended_series"`
}
