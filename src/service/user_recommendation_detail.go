package service

import (
	"context"
	"go-academy-drive-exercise/model"
)

type UserRecommendationDetailReader interface {
	GetCommonSeries(ctx context.Context, userID uint64, otherUserID uint64) ([]model.WatchSeries, error)
	GetCommonGenres(ctx context.Context, userID uint64, otherUserID uint64) ([]model.Genre, error)
	GetRecommendedSeries(ctx context.Context, userID uint64) ([]model.WatchSeries, error)
}

type UserRecommendationDetailService struct {
	userChecker                 UserChecker
	recommendedUserDetailReader UserRecommendationDetailReader
}

func NewUserRecommendationDetailService(userChecker UserChecker, recommendedUserDetailReader UserRecommendationDetailReader) *UserRecommendationDetailService {
	return &UserRecommendationDetailService{
		userChecker:                 userChecker,
		recommendedUserDetailReader: recommendedUserDetailReader,
	}
}

func (s *UserRecommendationDetailService) GetUserRecommendationDetail(ctx context.Context, userID uint64, otherUserID uint64) (*model.UserRecommendationDetail, error) {

	if userID == otherUserID {
		return nil, ErrSameUser
	}

	exists, err := s.userChecker.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	exists, err = s.userChecker.UserExists(ctx, otherUserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	commonSeries, err := s.recommendedUserDetailReader.GetCommonSeries(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}

	commonGenres, err := s.recommendedUserDetailReader.GetCommonGenres(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}

	recommendedSeries, err := s.recommendedUserDetailReader.GetRecommendedSeries(ctx, otherUserID)
	if err != nil {
		return nil, err
	}

	return &model.UserRecommendationDetail{
		UserID:            userID,
		OtherUserID:       otherUserID,
		CommonSeries:      commonSeries,
		CommonGenres:      commonGenres,
		RecommendedSeries: recommendedSeries,
	}, nil
}
