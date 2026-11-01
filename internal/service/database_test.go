package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// TestService_DatabasePing verifies DatabasePing returns nil when the
// database is reachable and an error otherwise.
func TestService_DatabasePing(t *testing.T) {
	t.Run("database reachable", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var nilRepository Repository
		mockDB := mocks.NewMockDBPinger(ctrl)
		svc := New(nilRepository, mockDB, *zap.NewNop())

		mockDB.EXPECT().PingContext(gomock.Any()).Return(nil)

		err := svc.DatabasePing(t.Context())
		assert.NoError(t, err)
	})

	t.Run("database unreachable", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var nilRepository Repository
		mockDB := mocks.NewMockDBPinger(ctrl)
		svc := New(nilRepository, mockDB, *zap.NewNop())

		mockDB.EXPECT().PingContext(gomock.Any()).Return(errors.New("connection refused"))

		err := svc.DatabasePing(t.Context())
		assert.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var nilRepository Repository
		mockDB := mocks.NewMockDBPinger(ctrl)
		svc := New(nilRepository, mockDB, *zap.NewNop())

		mockDB.EXPECT().PingContext(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			_, ok := ctx.Deadline()
			assert.True(t, ok)

			<-ctx.Done()
			return ctx.Err()
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		err := svc.DatabasePing(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})
	t.Run("context deadline exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var nilRepository Repository
		mockDB := mocks.NewMockDBPinger(ctrl)
		svc := New(nilRepository, mockDB, *zap.NewNop())

		mockDB.EXPECT().PingContext(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			_, ok := ctx.Deadline()
			assert.True(t, ok)

			<-ctx.Done()
			return ctx.Err()
		})

		err := svc.DatabasePing(t.Context())
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
