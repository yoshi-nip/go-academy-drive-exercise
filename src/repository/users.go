package repository

import (
	"context"
	"database/sql"
	"go-academy-drive-exercise/model"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	readDB *sql.DB
}

func NewUserRepository(readDB *sql.DB) *UserRepository {
	return &UserRepository{
		readDB: readDB,
	}
}

func (r *UserRepository) GetDemoUsers(ctx context.Context) ([]model.DemoUserID, error) {
	stmt, err := r.readDB.PrepareContext(ctx, "SELECT id FROM users ORDER BY RAND() LIMIT 5")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.DemoUserID
	for rows.Next() {
		var user model.DemoUserID
		if err := rows.Scan(&user.UserID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetWatchSeries(ctx context.Context, userID uint64) ([]model.WatchSeries, error) {
	stmt, err := r.readDB.PrepareContext(ctx, "SELECT s.id, s.title, s.genre_id "+
		"FROM user_watched_series AS w "+
		"JOIN series AS s ON s.id = w.series_id "+
		"WHERE w.user_id = ? "+
		"ORDER BY s.id "+
		"LIMIT 5;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	series := make([]model.WatchSeries, 0)

	for rows.Next() {
		var s model.WatchSeries
		if err := rows.Scan(&s.SeriesID, &s.Title, &s.GenreID); err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return series, nil
}
