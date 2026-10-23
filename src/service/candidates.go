package service

import (
	"context"
	"go-academy-drive-exercise/model"
)

type CandidateUserReader interface {
	GetSharedSeriesUsers(ctx context.Context, userID uint64) ([]model.UserID, error)
	GetSharedGenreUsers(ctx context.Context, userID uint64, limit int) ([]model.UserID, error)
	GetOtherUsers(ctx context.Context, userID uint64, limit int) ([]model.UserID, error)
}

type CandidateService struct {
	userChecker         UserChecker
	candidateUserReader CandidateUserReader
}

func NewCandidateService(userChecker UserChecker, candidateUserReader CandidateUserReader) *CandidateService {
	return &CandidateService{
		userChecker:         userChecker,
		candidateUserReader: candidateUserReader,
	}
}

func (s *CandidateService) GetCandidateUsers(ctx context.Context, userID uint64) ([]model.Candidate, error) {
	exists, err := s.userChecker.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	seriesCandidates, err := s.candidateUserReader.GetSharedSeriesUsers(ctx, userID)
	if err != nil {
		return nil, err
	}

	candidate := toCandidates(seriesCandidates, "shared_series")

	if len(candidate) >= 10 {
		return candidate, nil
	}

	remaining := 10 - len(candidate)
	if remaining > 0 {
		genreCandidates, err := s.candidateUserReader.GetSharedGenreUsers(ctx, userID, remaining)
		if err != nil {
			return nil, err
		}
		candidate = append(candidate, toCandidates(genreCandidates, "shared_genre")...)
	}

	remaining = 10 - len(candidate)
	if remaining > 0 {
		otherCandidates, err := s.candidateUserReader.GetOtherUsers(ctx, userID, remaining)
		if err != nil {
			return nil, err
		}
		candidate = append(candidate, toCandidates(otherCandidates, "other")...)
	}

	return candidate, nil

}

func toCandidates(users []model.UserID, reason string) []model.Candidate {
	candidates := make([]model.Candidate, len(users))
	for i, user := range users {
		candidates[i] = model.Candidate{
			UserID: user.UserID,
			Reason: reason,
		}
	}
	return candidates
}
