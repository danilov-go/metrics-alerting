package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danilov-go/metrics-alerting.git/internal/config/db"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("error test")

func Test_storageDB_SaveCounters(t *testing.T) {
	query := `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
		`
	tests := []struct {
		name      string
		mName     string
		delta     int64
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:  "положительный тест",
			mName: "PollCount",
			delta: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("PollCount", int64(5)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name:  "ошибка базы данных",
			mName: "PollCount",
			delta: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("PollCount", int64(5)).
					WillReturnError(errTest)
			},
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			err = storage.SaveCounters(context.Background(), tt.mName, tt.delta)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_SaveGauges(t *testing.T) {
	query := `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`
	tests := []struct {
		name      string
		mName     string
		value     float64
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:  "положительный тест",
			mName: "Alloc",
			value: 123.45,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("Alloc", 123.45).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name:  "ошибка базы данных",
			mName: "Alloc",
			value: 123.45,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(query).
					WithArgs("Alloc", 123.45).
					WillReturnError(errTest)
			},
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			err = storage.SaveGauges(context.Background(), tt.mName, tt.value)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetGauges(t *testing.T) {
	query := `SELECT value FROM metrics WHERE mtype = 'gauge' AND id = $1`
	tests := []struct {
		name      string
		mName     string
		setupMock func(mock sqlmock.Sqlmock)
		wantValue float64
		wantErr   error
	}{
		{
			name:  "положительный тест",
			mName: "Alloc",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value"}).AddRow(123.45)
				mock.ExpectQuery(query).WithArgs("Alloc").WillReturnRows(rows)
			},
			wantValue: 123.45,
			wantErr:   nil,
		},
		{
			name:  "ошибка базы данных",
			mName: "Alloc",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).WithArgs("Alloc").WillReturnError(errTest)
			},
			wantValue: 0,
			wantErr:   errTest,
		},
		{
			name:  "база вернула NULL",
			mName: "Alloc",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value"}).AddRow(nil)
				mock.ExpectQuery(query).WithArgs("Alloc").WillReturnRows(rows)
			},
			wantValue: 0,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			val, err := storage.GetGauges(context.Background(), tt.mName)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantValue, val)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetCounters(t *testing.T) {
	query := `SELECT delta FROM metrics WHERE mtype = 'counter' AND id = $1`
	tests := []struct {
		name      string
		mName     string
		setupMock func(mock sqlmock.Sqlmock)
		wantDelta int64
		wantErr   error
	}{
		{
			name:  "положительный тест",
			mName: "PollCount",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"delta"}).AddRow(10)
				mock.ExpectQuery(query).WithArgs("PollCount").WillReturnRows(rows)
			},
			wantDelta: 10,
			wantErr:   nil,
		},
		{
			name:  "ошибка базы данных",
			mName: "PollCount",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).WithArgs("PollCount").WillReturnError(errTest)
			},
			wantDelta: 0,
			wantErr:   errTest,
		},
		{
			name:  "база вернула NULL",
			mName: "PollCount",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"delta"}).AddRow(nil)
				mock.ExpectQuery(query).WithArgs("PollCount").WillReturnRows(rows)
			},
			wantDelta: 0,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			val, err := storage.GetCounters(context.Background(), tt.mName)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantDelta, val)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetAllGauges(t *testing.T) {
	query := `SELECT id, value FROM metrics WHERE mtype = 'gauge'`
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		want      map[string]float64
		wantErr   error
	}{
		{
			name: "положительный тест",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "value"}).
					AddRow("Alloc", 123.45).
					AddRow("CPUutilization1", 38.88).
					AddRow("CPUutilization2", 35.17).
					AddRow("CPUutilization3", 22.22).
					AddRow("CPUutilization4", 18.50).
					AddRow("CPUutilization5", 0.99).
					AddRow("CPUutilization6", 0.50).
					AddRow("CPUutilization7", 0.49).
					AddRow("CPUutilization8", nil)
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			want: map[string]float64{
				"Alloc":           123.45,
				"CPUutilization1": 38.88,
				"CPUutilization2": 35.17,
				"CPUutilization3": 22.22,
				"CPUutilization4": 18.50,
				"CPUutilization5": 0.99,
				"CPUutilization6": 0.50,
				"CPUutilization7": 0.49,
			},
			wantErr: nil,
		},
		{
			name: "ошибка QueryContext",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).WillReturnError(errTest)
			},
			want:    nil,
			wantErr: errTest,
		},
		{
			name: "ошибка rows.Err",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "value"}).
					AddRow("Alloc", 12.34).
					RowError(0, errTest)
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			want:    nil,
			wantErr: errTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			res, err := storage.GetAllGauges(context.Background())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetAllCounters(t *testing.T) {
	query := `SELECT id, delta FROM metrics WHERE mtype = 'counter'`
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		want      map[string]int64
		wantErr   error
	}{
		{
			name: "успешное получение всех counter",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "delta"}).
					AddRow("PollCount", int64(5)).
					AddRow("PollCountTest", nil)
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			want: map[string]int64{
				"PollCount": 5,
			},
			wantErr: nil,
		},
		{
			name: "ошибка QueryContext",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).WillReturnError(errTest)
			},
			want:    nil,
			wantErr: errTest,
		},
		{
			name: "ошибка rows.Err во время итерации",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "delta"}).
					AddRow("PollCount", int64(5)).
					RowError(0, errTest)
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			want:    nil,
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			res, err := storage.GetAllCounters(context.Background())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_SaveAll(t *testing.T) {
	stmtCounterQuery := `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta
	`
	stmtGaugeQuery := `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value
		`
	deltaVal := int64(5)
	tests := []struct {
		name      string
		metrics   []models.Metrics
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "положительный тест",
			metrics: []models.Metrics{
				{ID: "Alloc", MType: models.Gauge, Value: models.PointerFloat64(123.45)},
				{ID: "CPUutilization1", MType: models.Gauge, Value: models.PointerFloat64(38.88)},
				{ID: "CPUutilization2", MType: models.Gauge, Value: models.PointerFloat64(35.17)},
				{ID: "CPUutilization3", MType: models.Gauge, Value: models.PointerFloat64(22.22)},
				{ID: "CPUutilization4", MType: models.Gauge, Value: models.PointerFloat64(18.50)},
				{ID: "CPUutilization5", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
				{ID: "CPUutilization6", MType: models.Gauge, Value: models.PointerFloat64(0.50)},
				{ID: "CPUutilization7", MType: models.Gauge, Value: models.PointerFloat64(0.49)},
				{ID: "CPUutilization8", MType: models.Gauge, Value: models.PointerFloat64(0.99)},
				{ID: "PollCount", MType: models.Counter, Delta: models.PointerInt64(5)},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare(stmtCounterQuery)
				mock.ExpectPrepare(stmtGaugeQuery)
				mock.ExpectExec(stmtGaugeQuery).WithArgs("Alloc", 123.45).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization1", 38.88).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization2", 35.17).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization3", 22.22).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization4", 18.50).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization5", 0.99).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization6", 0.50).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization7", 0.49).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtGaugeQuery).WithArgs("CPUutilization8", 0.99).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(stmtCounterQuery).WithArgs("PollCount", int64(5)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "ошибка начала транзакции",
			metrics: []models.Metrics{
				{ID: "PollCount", MType: models.Counter, Delta: &deltaVal},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errTest)
			},
			wantErr: true,
		},
		{
			name: "delta равна nil",
			metrics: []models.Metrics{
				{ID: "PollCount", MType: models.Counter, Delta: nil},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare(stmtCounterQuery)
				mock.ExpectPrepare(stmtGaugeQuery)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			err = storage.SaveAll(context.Background(), tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_Ping(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "положительный тест",
			wantErr: nil,
		},
		{
			name:    "ошибка базы данных",
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer func() { _ = sql.Close() }()
			exp := mock.ExpectPing()
			if tt.wantErr != nil {
				exp.WillReturnError(tt.wantErr)
			}
			storage := db.NewStorageDB(sql)
			err = storage.Ping(t.Context())
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
