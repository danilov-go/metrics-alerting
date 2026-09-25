package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"go.uber.org/zap/zaptest"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestGRPCSender(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	logger := zaptest.NewLogger(t)
	gaugeVal := 123.45
	counterDelta := int64(5)
	metrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &gaugeVal},
		{ID: "PollCount", MType: models.Counter, Delta: &counterDelta},
	}
	testHost := "10.0.0.1"
	testInterceptor := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		assert.Contains(t, method, "DecryptedMetrics")
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, testHost, md.Get("x-real-ip")[0])
		cipherReq, ok := req.(*pb.EncryptedMetricsRequest)
		require.True(t, ok)
		body := cipherReq.GetCipher()
		cryptoKeyHex := md.Get("crypto-key")
		require.NotEmpty(t, cryptoKeyHex)
		cipherKey, err := base64.StdEncoding.DecodeString(cryptoKeyHex[0])
		require.NoError(t, err)
		body, err = crypto.DecryptBody(cipherKey, body, privateKey)
		require.NoError(t, err)
		var updateReq pb.UpdateMetricsRequest
		err = proto.Unmarshal(body, &updateReq)
		require.NoError(t, err)
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
	sender, err := NewGRPCSender(conn, testHost, logger.Sugar())
	defer func() {
		err := sender.Close()
		assert.NoError(t, err)
		err = sender.Close()
		assert.Error(t, err)
	}()
	require.NoError(t, err)
	sender.client = pb.NewMetricsClient(conn)
	sender.conn = conn
	sender.Send(context.Background(), metrics, publicKey)
}
