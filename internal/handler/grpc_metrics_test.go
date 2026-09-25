package handler_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestGRPCHandler_UpdateMetrics(t *testing.T) {
	nameGauge := "Alloc"
	nameCounter := "PollCount"
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
			h := handler.NewGRPCHandler(tt.storage, logger.Sugar(), nil)
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

func TestGRPCHandler_DecryptedMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	nameGauge := "Alloc"
	nameCounter := "PollCount"
	value := 123.45
	delta := int64(5)
	metrics := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{Id: nameGauge, Type: pb.Metric_GAUGE, Value: value},
			{Id: nameCounter, Type: pb.Metric_COUNTER, Delta: delta},
		},
	}
	rawBytes, _ := proto.Marshal(metrics)
	cipherBody, cipherKey, _ := crypto.EncryptBody(&privKey.PublicKey, rawBytes)
	encodedKey := base64.StdEncoding.EncodeToString(cipherKey)
	tests := []struct {
		name      string
		useCrypto bool
		key       string
		cipher    []byte
		code      codes.Code
	}{
		{
			name:      "положительный тест",
			useCrypto: true,
			key:       encodedKey,
			cipher:    cipherBody,
			code:      codes.OK,
		},
		{
			name:      "отсутствует crypto-key",
			useCrypto: true,
			key:       "",
			cipher:    cipherBody,
			code:      codes.InvalidArgument,
		},
		{
			name:      "ошибка расшифровки",
			useCrypto: true,
			key:       encodedKey,
			cipher:    []byte("test"),
			code:      codes.InvalidArgument,
		},
		{
			name:      "без шифрования",
			useCrypto: false,
			key:       "",
			cipher:    rawBytes,
			code:      codes.OK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &repository.MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
			var key *rsa.PrivateKey
			if tt.useCrypto {
				key = privKey
			}
			h := handler.NewGRPCHandler(storage, logger.Sugar(), key)
			ctx := context.Background()
			if tt.key != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("crypto-key", tt.key))
			}
			resp, err := h.DecryptedMetrics(ctx, &pb.EncryptedMetricsRequest{Cipher: tt.cipher})
			st, _ := status.FromError(err)
			assert.Equal(t, tt.code, st.Code())
			if tt.code == codes.OK {
				assert.NotNil(t, resp)
				assert.Equal(t, value, storage.Gauges[nameGauge])
				assert.Equal(t, delta, storage.Counters[nameCounter])
			}
		})
	}
}
