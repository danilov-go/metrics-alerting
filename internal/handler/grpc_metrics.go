package handler

import (
	"context"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler связывает GRPC-запросов с хранилищем данных.
type GRPCHandler struct {
	pb.UnimplementedMetricsServer
	storage Storage
	logger  log
}

// NewGRPCHandler создает новый экземпляр Handler.
func NewGRPCHandler(storage Storage, l log) *GRPCHandler {
	return &GRPCHandler{
		storage: storage,
		logger:  l,
	}
}

// UpdateMetrics обрабатывает входящий батч метрик от агента.
func (h *GRPCHandler) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	pbMetrics := req.GetMetrics()
	if len(pbMetrics) == 0 {
		return &pb.UpdateMetricsResponse{}, nil
	}
	metrics := make([]models.Metrics, 0, len(pbMetrics))
	for _, m := range pbMetrics {
		if m.GetId() == "" {
			h.logger.Errorw("не задан ID метрики", "mType", m.GetType())
			return nil, status.Error(codes.InvalidArgument, "не задан ID метрики")
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
			return nil, status.Error(codes.InvalidArgument, "неизвестный тип метрики")
		}
	}
	if err := h.storage.SaveAll(ctx, metrics); err != nil {
		h.logger.Errorw("ошибка сохранения метрик", "error", err)
		return nil, status.Errorf(codes.Internal, "внутренняя ошибка сервера")
	}
	return &pb.UpdateMetricsResponse{}, nil
}
