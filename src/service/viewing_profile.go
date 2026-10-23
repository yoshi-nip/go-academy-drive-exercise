package service

import (
	"context"
	"errors"
	"go-academy-drive-exercise/model"
)

type UserChecker interface {
	UserExists(ctx context.Context, userID uint64) (bool, error)
}
type WatchSeriesReader interface {
	GetWatchSeries(ctx context.Context, userID uint64) ([]model.WatchSeries, error)
}
type WatchGenreReader interface {
	GetWatchGenreCounts(ctx context.Context, userID uint64) ([]model.GenreWatchCount, error)
}

type ViewingProfileService struct {
	userChecker       UserChecker       // DBでチェックする、repo層
	watchSeriesReader WatchSeriesReader // 視聴シリーズを取得する、repo層
	watchGenreReader  WatchGenreReader  // 視聴ジャンルを取得する、repo層
}

var ErrUserNotFound = errors.New("ユーザーが見つかりませんでした。s")

func NewViewingProfileService(userChecker UserChecker, watchSeriesReader WatchSeriesReader, watchGenreReader WatchGenreReader) *ViewingProfileService {
	return &ViewingProfileService{
		userChecker:       userChecker,
		watchSeriesReader: watchSeriesReader,
		watchGenreReader:  watchGenreReader,
	}
}

func (s *ViewingProfileService) GetViewingProfile(ctx context.Context, userID uint64) (model.ViewingProfile, error) {
	exists, err := s.userChecker.UserExists(ctx, userID)
	if err != nil {
		return model.ViewingProfile{}, err
	}
	if !exists {
		return model.ViewingProfile{}, ErrUserNotFound
	}

	watchSeries, err := s.watchSeriesReader.GetWatchSeries(ctx, userID)
	if err != nil {
		return model.ViewingProfile{}, err
	}

	watchGenreCounts, err := s.watchGenreReader.GetWatchGenreCounts(ctx, userID)
	if err != nil {
		return model.ViewingProfile{}, err
	}

	return model.ViewingProfile{
		UserID: userID,
		Series: watchSeries,
		Genres: watchGenreCounts,
	}, nil
}
