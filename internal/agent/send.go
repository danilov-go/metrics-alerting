package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-resty/resty/v2"
)

func (a *Agent) send(ctx context.Context, metrics []models.Metrics) {
	jsonMetric, err := json.Marshal(metrics)
	if err != nil {
		a.logger.Errorw("ошибка сериализации", "err", err)
		return
	}
	var hash string
	if a.key != "" {
		h := hmac.New(sha256.New, []byte(a.key))
		h.Write(jsonMetric)
		hash = hex.EncodeToString(h.Sum(nil))
	}
	var buf bytes.Buffer
	wg := gzip.NewWriter(&buf)
	_, err = wg.Write(jsonMetric)
	if err != nil {
		a.logger.Errorw("ошибка сжатия данных", "err", err)
		if errClose := wg.Close(); errClose != nil {
			a.logger.Errorw("ошибка закрытия gzip writer", "error", errClose)
		}
		return
	}
	err = wg.Close()
	if err != nil {
		a.logger.Errorw("ошибка закрытия gzip writer", "error", err)
		return
	}
	body := buf.Bytes()
	const maxRetries = 3
	duration := 1
	var response *resty.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		res := a.Client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(body)
		if a.key != "" {
			res.SetHeader("HashSHA256", hash)
		}
		response, err = res.Post("/updates/")
		if err == nil {
			break
		}
		if attempt == maxRetries {
			a.logger.Errorw("попытки отправки исчерпаны", "maxRetries", maxRetries)
			return
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			timer := time.NewTimer(time.Duration(duration) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			timer.Stop()
			duration += 2
			continue
		}
	}
	if response == nil {
		a.logger.Errorw("не удалось получить ответ от сервера", "error", err)
		return
	}
	if response.StatusCode() != http.StatusOK {
		a.logger.Errorw("статус запроса:", "status", response.StatusCode())
		return
	}
}
