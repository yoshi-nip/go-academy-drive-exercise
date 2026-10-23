package service

import (
	"context"
	"go-academy-drive-exercise/model"
)

type CandidateUserReader interface {
	GetSharedSeriesUsers(ctx context.Context, userID uint64) ([]model.UserID, error)
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
	return s.candidateUserReader.GetSharedSeriesUsers(ctx, userID)

}
