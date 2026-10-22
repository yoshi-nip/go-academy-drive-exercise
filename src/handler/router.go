package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Users *UserHandler
}

func NewRouter(h Handlers) *gin.Engine {
	r := gin.Default()
	r.GET("/health", HealthHandler)
	r.GET("/demo-users", h.Users.GetDemoUsersHandler())
	r.GET("/users/:userID/watch-series", h.Users.GetWatchSeriesHandler())
	r.GET("/users/:userID/watch-genre-counts", h.Users.GetWatchGenreCountsHandler())
	return r
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
