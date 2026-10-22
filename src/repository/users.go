package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DemoUserID struct {
	UserID int
}

type UserRepository struct {
	readDB *sql.DB
}

func NewUserRepository(readDB *sql.DB) *UserRepository {
	return &UserRepository{
		readDB: readDB,
	}
}

func (r *UserRepository) GetDemoUsers(ctx context.Context) ([]DemoUserID, error) {
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

	var users []DemoUserID
	for rows.Next() {
		var user DemoUserID
		if err := rows.Scan(&user.UserID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	fmt.Println(users) // デバッグ用に取得したユーザーを表示する場合
	return users, nil
}
