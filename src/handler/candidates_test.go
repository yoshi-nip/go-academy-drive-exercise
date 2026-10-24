package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
	"go-academy-drive-exercise/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type candidateServiceStub struct {
	get func(context.Context, uint64) ([]model.Candidate, error)
}

func (s candidateServiceStub) GetCandidateUsers(c context.Context, id uint64) ([]model.Candidate, error) {
	return s.get(c, id)
}
func TestGetCandidateUsersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, id      string
		users         []model.Candidate
		err           error
		status, calls int
		body          string
	}{
		{name: "success", id: "42", users: []model.Candidate{{UserID: 7, Reason: "shared_series"}, {UserID: 8, Reason: "shared_genre"}, {UserID: 9, Reason: "other"}}, status: 200, calls: 1, body: `{"users":[{"user_id":7,"reason":"shared_series"},{"user_id":8,"reason":"shared_genre"},{"user_id":9,"reason":"other"}]}`},
		{name: "empty", id: "42", users: []model.Candidate{}, status: 200, calls: 1, body: `{"users":[]}`},
		{name: "missing", id: "42", err: fmt.Errorf("lookup: %w", service.ErrUserNotFound), status: 404, calls: 1, body: `{"error":"ユーザーが存在しません"}`},
		{name: "error", id: "42", users: []model.Candidate{{UserID: 7, Reason: "shared_series"}}, err: errors.New("private detail"), status: 500, calls: 1, body: `{"error":"候補ユーザーの取得に失敗しました"}`},
		{name: "letters", id: "abc", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "zero", id: "0", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "negative", id: "-1", status: 400, body: `{"error":"userIDが不正です"}`},
		{name: "overflow", id: "18446744073709551616", status: 400, body: `{"error":"userIDが不正です"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			source := candidateServiceStub{get: func(c context.Context, id uint64) ([]model.Candidate, error) {
				calls++
				require.Equal(t, ctx, c)
				require.Equal(t, uint64(42), id)
				return tc.users, tc.err
			}}
			router := NewRouter(Handlers{Users: NewUserHandler(nil, nil, nil, nil, source, nil)})
			req := httptest.NewRequest(http.MethodGet, "/users/"+tc.id+"/candidate-users", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			require.JSONEq(t, tc.body, w.Body.String())
			require.Equal(t, tc.calls, calls)
		})
	}
}
