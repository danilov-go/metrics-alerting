package agent

import (
	"context"
	"testing"

	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGRPCSender(t *testing.T) {
	logger := zaptest.NewLogger(t)
	gaugeVal := 123.45
	counterDelta := int64(5)
	metrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &gaugeVal},
		{ID: "PollCount", MType: models.Counter, Delta: &counterDelta},
	}
	testHost := "10.0.0.1"
	testInterceptor := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		assert.Contains(t, method, "UpdateMetrics")
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, testHost, md.Get("x-real-ip")[0])
		updateReq, ok := req.(*pb.UpdateMetricsRequest)
		require.True(t, ok)
		require.Len(t, updateReq.Metrics, 2)
		assert.Equal(t, "Alloc", updateReq.Metrics[0].Id)
		assert.Equal(t, pb.Metric_GAUGE, updateReq.Metrics[0].Type)
		assert.Equal(t, gaugeVal, updateReq.Metrics[0].Value)
		assert.Equal(t, "PollCount", updateReq.Metrics[1].Id)
		assert.Equal(t, pb.Metric_COUNTER, updateReq.Metrics[1].Type)
		assert.Equal(t, counterDelta, updateReq.Metrics[1].Delta)
		return nil
	}
	conn, err := grpc.NewClient(testHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(testInterceptor),
	)
	require.NoError(t, err)
	defer conn.Close()
	sender, err := NewGRPCSender(testHost, testHost, logger.Sugar())
	require.NoError(t, err)
	sender.client = pb.NewMetricsClient(conn)
	sender.conn = conn
	sender.Send(context.Background(), metrics, nil)
	err = sender.Close()
	assert.NoError(t, err)
}
