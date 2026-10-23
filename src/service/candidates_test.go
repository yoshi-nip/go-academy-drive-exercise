package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"go-academy-drive-exercise/model"
	"testing"
)

type candidateSourceStub struct {
	check func(context.Context, uint64) (bool, error)
	get   func(context.Context, uint64) ([]model.UserID, error)
}

func (s candidateSourceStub) UserExists(c context.Context, id uint64) (bool, error) {
	return s.check(c, id)
}
func (s candidateSourceStub) GetSharedSeriesUsers(c context.Context, id uint64) ([]model.UserID, error) {
	return s.get(c, id)
}
func TestGetCandidateUsers(t *testing.T) {
	failure := errors.New("read failed")
	for _, tc := range []struct {
		name                        string
		exists                      bool
		checkErr, getErr, errorWant error
		users                       []model.UserID
		calls                       int
	}{
		{name: "success", exists: true, users: []model.UserID{{UserID: 7}}, calls: 1},
		{name: "empty", exists: true, users: []model.UserID{}, calls: 1},
		{name: "missing", errorWant: ErrUserNotFound},
		{name: "check_error", checkErr: failure, errorWant: failure},
		{name: "get_error", exists: true, getErr: failure, errorWant: failure, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			checked := false
			calls := 0
			source := candidateSourceStub{
				check: func(c context.Context, id uint64) (bool, error) {
					require.Equal(t, ctx, c)
					require.Equal(t, uint64(42), id)
					checked = true
					return tc.exists, tc.checkErr
				},
				get: func(c context.Context, id uint64) ([]model.UserID, error) {
					require.True(t, checked)
					require.Equal(t, ctx, c)
					require.Equal(t, uint64(42), id)
					calls++
					return tc.users, tc.getErr
				},
			}
			got, err := NewCandidateService(source, source).GetCandidateUsers(ctx, 42)
			require.True(t, checked)
			require.Equal(t, tc.calls, calls)
			if tc.errorWant != nil {
				require.ErrorIs(t, err, tc.errorWant)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.users, got)
			}
		})
	}
}
