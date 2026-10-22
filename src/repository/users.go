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

func GetDemoUsers(ctx context.Context, db *sql.DB) ([]DemoUserID, error) {
	stmt, err := db.PrepareContext(ctx, "SELECT id FROM users ORDER BY RAND() LIMIT 5")
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
