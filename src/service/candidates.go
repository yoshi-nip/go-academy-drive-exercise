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

func (s *CandidateService) GetCandidateUsers(ctx context.Context, userID uint64) ([]model.UserID, error) {
	exists, err := s.userChecker.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	candidate, err := s.candidateUserReader.GetSharedSeriesUsers(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(candidate) >= 10 {
		return candidate, nil
	}

	remaining := 10 - len(candidate)
	if remaining > 0 {
		genreCandidates, err := s.candidateUserReader.GetSharedGenreUsers(ctx, userID, remaining)
		if err != nil {
			return nil, err
		}
		candidate = append(candidate, genreCandidates...)
	}

	remaining = 10 - len(candidate)
	if remaining > 0 {
		otherCandidates, err := s.candidateUserReader.GetOtherUsers(ctx, userID, remaining)
		if err != nil {
			return nil, err
		}
		candidate = append(candidate, otherCandidates...)
	}

	return candidate, nil

}
