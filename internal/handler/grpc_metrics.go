package handler

import (
	"context"
	"crypto/rsa"
	"encoding/base64"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// GRPCHandler связывает GRPC-запросов с хранилищем данных.
type GRPCHandler struct {
	pb.UnimplementedMetricsServer
	storage    Storage
	logger     log
	privateKey *rsa.PrivateKey
}

// NewGRPCHandler создает новый экземпляр Handler.
func NewGRPCHandler(storage Storage, l log, key *rsa.PrivateKey) *GRPCHandler {
	return &GRPCHandler{
		storage:    storage,
		logger:     l,
		privateKey: key,
	}
}

// UpdateMetrics обрабатывает входящий батч метрик от агента.
func (h *GRPCHandler) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if err := h.save(ctx, req.GetMetrics()); err != nil {
		return nil, err
	}
	return &pb.UpdateMetricsResponse{}, nil
}

// DecryptedMetrics расшифровывает тело запроса и сохраняет содержащиеся в нем метрики.
func (h *GRPCHandler) DecryptedMetrics(ctx context.Context, req *pb.EncryptedMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	cipherBody := req.GetCipher()
	var cifer []byte
	if h.privateKey != nil {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("crypto-key")) == 0 {
			return nil, status.Error(codes.InvalidArgument, "в метаданных отсутствует ключ")
		}
		strKey := md.Get("crypto-key")[0]
		cipherKey, err := base64.StdEncoding.DecodeString(strKey)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "некорректный формат ключа в метаданных")
		}
		var errDecrypt error
		cifer, errDecrypt = crypto.DecryptBody(cipherKey, cipherBody, h.privateKey)
		if errDecrypt != nil {
			h.logger.Errorw("ошибка расшифровки", "error", errDecrypt)
			return nil, status.Error(codes.InvalidArgument, "ошибка расшифровки")
		}
	} else {
		cifer = cipherBody
	}
	var metricsReq pb.UpdateMetricsRequest
	if err := proto.Unmarshal(cifer, &metricsReq); err != nil {
		h.logger.Errorw("ошибка десериализации", "error", err)
		return nil, status.Error(codes.InvalidArgument, "ошибка десериализации")
	}
	if err := h.save(ctx, metricsReq.GetMetrics()); err != nil {
		return nil, err
	}
	return &pb.UpdateMetricsResponse{}, nil
}

func (h *GRPCHandler) save(ctx context.Context, pbMetrics []*pb.Metric) error {
	if len(pbMetrics) == 0 {
		return nil
	}
	metrics := make([]models.Metrics, 0, len(pbMetrics))
	for _, m := range pbMetrics {
		if m.GetId() == "" {
			h.logger.Errorw("не задан ID метрики", "mType", m.GetType())
			return status.Error(codes.InvalidArgument, "не задан ID метрики")
		}
		switch m.GetType() {
		case pb.Metric_GAUGE:
			value := m.GetValue()
			metric := models.Metrics{
				ID:    m.GetId(),
				MType: models.Gauge,
				Value: &value,
			}
			metrics = append(metrics, metric)
		case pb.Metric_COUNTER:
			delta := m.GetDelta()
			metric := models.Metrics{
				ID:    m.GetId(),
				MType: models.Counter,
				Delta: &delta,
			}
			metrics = append(metrics, metric)
		default:
			h.logger.Errorw("неизвестный тип метрики", "mType", m.GetId())
			return status.Error(codes.InvalidArgument, "неизвестный тип метрики")
		}
	}
	if err := h.storage.SaveAll(ctx, metrics); err != nil {
		h.logger.Errorw("ошибка сохранения метрик", "error", err)
		return status.Errorf(codes.Internal, "внутренняя ошибка сервера")
	}
	return nil
}
