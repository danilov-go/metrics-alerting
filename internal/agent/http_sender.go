package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/go-resty/resty/v2"
)

// HTTPSender отправляет метрики на сервер по протоколу HTTP.
type HTTPSender struct {
	client *resty.Client
	logger log
	key    string
	host   string
}

// NewHTTPSender создает экзепляр HTTPSender.
func NewHTTPSender(serverURL string, key string, hostIP string, l log) *HTTPSender {
	client := resty.New().
		SetTimeout(time.Second * 1).
		SetBaseURL("http://" + serverURL).
		SetRetryCount(3).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			return time.Duration(1+2*(r.Request.Attempt-1)) * time.Second, nil
		})
	return &HTTPSender{
		client: client,
		logger: l,
		key:    key,
		host:   hostIP,
	}
}

// Send преобразует и отправляет метрики на HTTP-сервер.
func (s *HTTPSender) Send(ctx context.Context, metrics []models.Metrics, publicKey *rsa.PublicKey) {
	jsonMetric, err := json.Marshal(metrics)
	if err != nil {
		s.logger.Errorw("ошибка сериализации", "err", err)
		return
	}
	var hash string
	if s.key != "" {
		h := hmac.New(sha256.New, []byte(s.key))
		h.Write(jsonMetric)
		hash = hex.EncodeToString(h.Sum(nil))
	}
	var buf bytes.Buffer
	wg := gzip.NewWriter(&buf)
	_, err = wg.Write(jsonMetric)
	if err != nil {
		s.logger.Errorw("ошибка сжатия данных", "err", err)
		if errClose := wg.Close(); errClose != nil {
			s.logger.Errorw("ошибка закрытия gzip writer", "error", errClose)
		}
		return
	}
	err = wg.Close()
	if err != nil {
		s.logger.Errorw("ошибка закрытия gzip writer", "error", err)
		return
	}
	body := buf.Bytes()
	var cipherKey []byte
	var cipherBody []byte
	if publicKey != nil {
		cipherBody, cipherKey, err = encrypt(publicKey, body)
		if err != nil {
			s.logger.Errorw("ошибка шифрования", "error", err)
			return
		}
		body = cipherBody
	}
	req := s.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body)

	if s.key != "" {
		req.SetHeader("HashSHA256", hash)
	}
	if publicKey != nil {
		req.SetHeader("Crypto-Key", hex.EncodeToString(cipherKey))
	}
	if s.host != "" {
		req.SetHeader("X-Real-IP", s.host)
	}
	response, err := req.Post("/updates/")
	if err != nil {
		s.logger.Errorw("попытки отправки исчерпаны", "error", err)
		return
	}
	if response == nil {
		s.logger.Errorw("не удалось получить ответ от сервера: response равен nil")
		return
	}
	if response.StatusCode() != http.StatusOK {
		s.logger.Errorw("статус запроса:", "status", response.StatusCode())
		return
	}
}

func encrypt(publicKey *rsa.PublicKey, body []byte) ([]byte, []byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, err
	}
	gsm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gsm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	cipherBody := gsm.Seal(nonce, nonce, body, nil)
	cipherKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, nil, err
	}
	return cipherBody, cipherKey, nil
}
