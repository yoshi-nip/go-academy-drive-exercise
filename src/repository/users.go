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

func (r *UserRepository) GetDemoUsers(ctx context.Context) ([]model.UserID, error) {
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

	var users []model.UserID
	for rows.Next() {
		var user model.UserID
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
	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT s.id, s.title, s.genre_id "+
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

func (r *UserRepository) GetCommonSeries(ctx context.Context, userID uint64, otherUserID uint64) ([]model.WatchSeries, error) {
	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT s.id AS series_id, s.title, s.genre_id "+
			"FROM user_watched_series AS mine "+
			"JOIN user_watched_series AS other "+
			"    ON other.series_id = mine.series_id "+
			"JOIN series AS s "+
			"    ON s.id = mine.series_id "+
			"WHERE mine.user_id = ? "+
			"  AND other.user_id = ? "+
			"ORDER BY s.id;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID, otherUserID)
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

func (r *UserRepository) GetWatchGenreCounts(ctx context.Context, userID uint64) ([]model.GenreWatchCount, error) {
	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT s.genre_id, COUNT(*) AS series_count "+
			"FROM user_watched_series AS w "+
			"JOIN series AS s ON s.id = w.series_id "+
			"WHERE w.user_id = ? "+
			"  AND s.genre_id <> '' "+
			"GROUP BY s.genre_id "+
			"ORDER BY series_count DESC, s.genre_id ASC;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	genreCounts := make([]model.GenreWatchCount, 0)

	for rows.Next() {
		var g model.GenreWatchCount
		if err := rows.Scan(&g.GenreID, &g.SeriesCount); err != nil {
			return nil, err
		}
		genreCounts = append(genreCounts, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genreCounts, nil
}

func (r *UserRepository) UserExists(ctx context.Context, userID uint64) (bool, error) {

	stmt, err := r.readDB.PrepareContext(ctx, "SELECT EXISTS( SELECT 1 FROM users WHERE id = ?)")

	if err != nil {
		return false, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var exists bool

	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, err
		}
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) GetSharedSeriesUsers(ctx context.Context, userID uint64) ([]model.UserID, error) {

	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT DISTINCT other.user_id "+
			"FROM user_watched_series AS mine "+
			"JOIN user_watched_series AS other "+
			"    ON mine.series_id = other.series_id "+
			"WHERE mine.user_id = ? "+
			"  AND other.user_id <> mine.user_id "+
			"ORDER BY RAND() "+
			"LIMIT 10;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]model.UserID, 0)

	for rows.Next() {
		var u model.UserID
		if err := rows.Scan(&u.UserID); err != nil {
			return nil, err
		}
		candidates = append(candidates, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

func (r *UserRepository) GetSharedGenreUsers(ctx context.Context, userID uint64, limit int) ([]model.UserID, error) {

	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT DISTINCT other.user_id "+
			"FROM user_watched_series AS mine "+
			"JOIN series AS my_series "+
			"    ON my_series.id = mine.series_id "+
			"JOIN series AS other_series "+
			"    ON other_series.genre_id = my_series.genre_id "+
			"JOIN user_watched_series AS other "+
			"    ON other.series_id = other_series.id "+
			"WHERE mine.user_id = ? "+
			"  AND my_series.genre_id <> '' "+
			"  AND other.user_id <> mine.user_id "+
			"  AND NOT EXISTS ( "+
			"      SELECT 1 "+
			"      FROM user_watched_series AS my_watch "+
			"      JOIN user_watched_series AS other_watch "+
			"          ON other_watch.series_id = my_watch.series_id "+
			"      WHERE my_watch.user_id = mine.user_id "+
			"        AND other_watch.user_id = other.user_id "+
			"  ) "+
			"ORDER BY RAND() "+
			"LIMIT ?;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]model.UserID, 0)

	for rows.Next() {
		var u model.UserID
		if err := rows.Scan(&u.UserID); err != nil {
			return nil, err
		}
		candidates = append(candidates, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

func (r *UserRepository) GetOtherUsers(ctx context.Context, userID uint64, limit int) ([]model.UserID, error) {

	stmt, err := r.readDB.PrepareContext(ctx,
		"SELECT u.id AS user_id "+
			"FROM users AS u "+
			"WHERE u.id <> ? "+
			"  AND EXISTS ( "+
			"      SELECT 1 "+
			"      FROM user_watched_series AS watched "+
			"      WHERE watched.user_id = u.id "+
			"  ) "+
			"  AND NOT EXISTS ( "+
			"      SELECT 1 "+
			"      FROM user_watched_series AS mine "+
			"      JOIN user_watched_series AS other "+
			"          ON other.series_id = mine.series_id "+
			"      WHERE mine.user_id = ? "+
			"        AND other.user_id = u.id "+
			"  ) "+
			"  AND NOT EXISTS ( "+
			"      SELECT 1 "+
			"      FROM user_watched_series AS mine "+
			"      JOIN series AS my_series "+
			"          ON my_series.id = mine.series_id "+
			"      JOIN series AS other_series "+
			"          ON other_series.genre_id = my_series.genre_id "+
			"      JOIN user_watched_series AS other "+
			"          ON other.series_id = other_series.id "+
			"      WHERE mine.user_id = ? "+
			"        AND my_series.genre_id <> '' "+
			"        AND other.user_id = u.id "+
			"  ) "+
			"ORDER BY RAND() "+
			"LIMIT ?;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]model.UserID, 0)

	for rows.Next() {
		var u model.UserID
		if err := rows.Scan(&u.UserID); err != nil {
			return nil, err
		}
		candidates = append(candidates, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}
