package agent

import (
	"context"
	"crypto/rsa"
	"encoding/base64"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	pb "github.com/danilov-go/metrics-alerting.git/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// GRPCSender отправляет метрики на сервер по протоколу gRPC.
type GRPCSender struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
	host   string
	logger log
}

// NewGRPCSender создает экзепляр GRPCSender.
func NewGRPCSender(conn *grpc.ClientConn, host string, l log) (*GRPCSender, error) {
	return &GRPCSender{
		client: pb.NewMetricsClient(conn),
		conn:   conn,
		host:   host,
		logger: l,
	}, nil
}

// Send преобразует и отправляет метрики на gRPC-сервер.
func (s *GRPCSender) Send(ctx context.Context, metrics []models.Metrics, publicKey *rsa.PublicKey) {
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
	md := metadata.New(map[string]string{"x-real-ip": s.host})
	var errReq error
	if publicKey != nil {
		rawBytes, err := proto.Marshal(req)
		if err != nil {
			s.logger.Errorw("ошибка сериализации", "error", err)
			return
		}
		cipherBody, cipherKey, err := crypto.EncryptBody(publicKey, rawBytes)
		if err != nil {
			s.logger.Errorw("ошибка шифрования", "error", err)
			return
		}
		encodedKey := base64.StdEncoding.EncodeToString(cipherKey)
		md.Set("crypto-key", encodedKey)
		ctx = metadata.NewOutgoingContext(ctx, md)
		cipherReq := &pb.EncryptedMetricsRequest{Cipher: cipherBody}
		_, errReq = s.client.DecryptedMetrics(ctx, cipherReq)
	} else {
		ctx = metadata.NewOutgoingContext(ctx, md)
		_, errReq = s.client.UpdateMetrics(ctx, req)
	}
	if errReq != nil {
		s.logger.Errorw("ошибка gRPC при обновлении метрик", "error", errReq)
		return
	}
}

func (s *GRPCSender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
