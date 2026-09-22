package agent

import (
	"context"
	"crypto/rsa"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCSender отправляет метрики на сервер по протоколу gRPC.
type GRPCSender struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
	host   string
	logger log
}

// NewGRPCSender создает экзепляр GRPCSender.
func NewGRPCSender(grpcAddress string, host string, l log) (*GRPCSender, error) {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCSender{
		client: pb.NewMetricsClient(conn),
		conn:   conn,
		host:   host,
		logger: l,
	}, nil
}

// Send преобразует и отправляет метрики на gRPC-сервер.
func (a *GRPCSender) Send(ctx context.Context, metrics []models.Metrics, publicKey *rsa.PublicKey) {
	var protoMetrics []*pb.Metric
	for _, m := range metrics {
		if m.ID == "" {
			continue
		}
		pm := &pb.Metric{Id: m.ID}
		switch m.MType {
		case models.Gauge:
			pm.Type = pb.Metric_GAUGE
			if m.Value != nil {
				pm.Value = *m.Value
			}
		case models.Counter:
			pm.Type = pb.Metric_COUNTER
			if m.Delta != nil {
				pm.Delta = *m.Delta
			}
		default:
			continue
		}
		protoMetrics = append(protoMetrics, pm)
	}
	req := &pb.UpdateMetricsRequest{Metrics: protoMetrics}
	md := metadata.New(map[string]string{"x-real-ip": a.host})
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err := a.client.UpdateMetrics(ctx, req)
	if err != nil {
		a.logger.Errorw("ошибка gRPC при обновлении метрик", "error", err)
		return
	}
}

func (a *GRPCSender) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}
