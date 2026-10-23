package model

type Candidate struct {
	UserID uint64 `json:"user_id"`
	Reason string `json:"reason"`
}
