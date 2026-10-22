package handler

import (
	"context"
	"go-academy-drive-exercise/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DemoUserReader interface {
	GetDemoUsers(ctx context.Context) ([]model.DemoUserID, error)
}

type WatchSeriesReader interface {
	GetWatchSeries(ctx context.Context, userID uint64) ([]model.WatchSeries, error)
}

type WatchGenreReader interface {
	GetWatchGenreCounts(
		ctx context.Context,
		userID uint64,
	) ([]model.GenreWatchCount, error)
}

type UserHandler struct {
	reader            DemoUserReader    // IF
	watchSeriesReader WatchSeriesReader // IF
	watchGenreReader  WatchGenreReader  // IF
}

// NewUserHandler は UserHandler のコンストラクタで reader(IF)を受け取る
func NewUserHandler(reader DemoUserReader, watchSeriesReader WatchSeriesReader, watchGenreReader WatchGenreReader) *UserHandler {
	return &UserHandler{
		reader:            reader,
		watchSeriesReader: watchSeriesReader,
		watchGenreReader:  watchGenreReader,
	}
}

func (h *UserHandler) GetDemoUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		users, err := h.reader.GetDemoUsers(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "demoユーザーの取得に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}

func (h *UserHandler) GetWatchSeriesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawUserId := c.Param("userID")
		userID, err := strconv.ParseUint(rawUserId, 10, 64)

		// 0も正常値と条件に含まれてしまうため条件に0も追加してはじく
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userIDが不正です"})
			return
		}
		series, err := h.watchSeriesReader.GetWatchSeries(ctx, uint64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "視聴シリーズの取得に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"series": series})
	}
}

func (h *UserHandler) GetWatchGenreCountsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawUserId := c.Param("userID")
		userID, err := strconv.ParseUint(rawUserId, 10, 64)

		// 0も正常値と条件に含まれてしまうため条件に0も追加してはじく
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userIDが不正です"})
			return
		}
		genreCounts, err := h.watchGenreReader.GetWatchGenreCounts(ctx, uint64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "視聴ジャンル集計の取得に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"genre": genreCounts})
	}
}
