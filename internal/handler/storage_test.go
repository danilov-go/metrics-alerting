package handler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type testStorage struct {
	count     int
	returnErr error
}

func (m *testStorage) SaveCounters(ctx context.Context, name string, delta int64) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) SaveGauges(ctx context.Context, name string, value float64) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) GetGauges(ctx context.Context, name string) (float64, error) {
	m.count++
	return 0, m.returnErr
}

func (m *testStorage) GetCounters(ctx context.Context, name string) (int64, error) {
	m.count++
	return 0, m.returnErr
}

func (m *testStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	m.count++
	return nil, m.returnErr
}

func (m *testStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	m.count++
	return nil, m.returnErr
}

func (m *testStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) Ping(ctx context.Context) error {
	m.count++
	return m.returnErr
}

func TestNewErrorMiddleware(t *testing.T) {
	retriableErr := &pgconn.PgError{
		Code: pgerrcode.ConnectionFailure,
	}
	nonRetriableErr := &pgconn.PgError{
		Code: pgerrcode.UniqueViolation,
	}
	expErr := fmt.Errorf("попытки подключения исчерпаны: %w", retriableErr)
	tests := []struct {
		name      string
		dbError   error
		wantCount int
		wantErr   error
	}{
		{
			name:      "с повтором",
			dbError:   retriableErr,
			wantCount: 4,
			wantErr:   expErr,
		},
		{
			name:      "без повтора",
			dbError:   nonRetriableErr,
			wantCount: 1,
			wantErr:   nonRetriableErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testStorage{returnErr: tt.dbError}
			middleware := handler.NewErrorMiddleware(h, 1*time.Millisecond, 0)
			check := func(err error, expCount int, expErr error) {
				t.Helper()
				assert.Equal(t, expCount, h.count)
				assert.Error(t, err)
				assert.Equal(t, expErr.Error(), err.Error())
				h.count = 0
			}
			ctx := context.Background()
			err := middleware.SaveCounters(ctx, "PollCount", 5)
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.SaveGauges(ctx, "Alloc", 123.45)
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetGauges(ctx, "Alloc")
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetCounters(ctx, "PollCount")
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.SaveAll(ctx, []models.Metrics{})
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.Ping(ctx)
			check(err, tt.wantCount, tt.wantErr)

		})
	}
}
