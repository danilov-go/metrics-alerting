package handler_test

import (
	"context"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCHandler_UpdateMetrics_WithMemStorage(t *testing.T) {
	nameGauge := "Alloc"
	nameCounter := "pollcount"
	value := 123.45
	delta := int64(5)
	tests := []struct {
		name       string
		req        *pb.UpdateMetricsRequest
		code       codes.Code
		storage    *repository.MemStorage
		expMessage string
	}{
		{
			name: "положительный тест",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{Id: nameGauge, Type: pb.Metric_GAUGE, Value: value},
					{Id: nameCounter, Type: pb.Metric_COUNTER, Delta: delta},
				},
			},
			code: codes.OK,
		},
		{
			name: "пустой запрос",
			req:  &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{}},
			code: codes.OK,
		},
		{
			name: "отсутствует ID",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{Id: "", Type: pb.Metric_GAUGE, Value: value},
				},
			},
			code:       codes.InvalidArgument,
			expMessage: "не задан ID метрики",
		},
		{
			name: "неверный тип метрики",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{Id: "test", Type: pb.Metric_MType(7)},
				},
			},
			code:       codes.InvalidArgument,
			expMessage: "неизвестный тип метрики",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			tt.storage = &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			h := handler.NewGRPCHandler(tt.storage, logger.Sugar())
			resp, err := h.UpdateMetrics(context.Background(), tt.req)
			if tt.code == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if len(tt.req.GetMetrics()) > 0 {
					assert.Equal(t, tt.storage.Counters[nameCounter], delta)
					assert.Equal(t, tt.storage.Gauges[nameGauge], value)
				}
			} else {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.code, st.Code())
				assert.Equal(t, tt.expMessage, st.Message())
				assert.Empty(t, tt.storage.Counters)
				assert.Empty(t, tt.storage.Gauges)
			}
		})
	}
}
