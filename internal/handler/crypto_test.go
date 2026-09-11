package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoMiddleware(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	expMetrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: models.Pointer(123.45)},
		{ID: "PollInterval", MType: models.Counter, Delta: models.Pointer(int64(5))},
	}
	expectedBody, err := json.Marshal(expMetrics)
	require.NoError(t, err)
	aesKey := make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, aesKey)
	require.NoError(t, err)
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	require.NoError(t, err)
	validHeader := hex.EncodeToString(encryptedAESKey)
	block, err := aes.NewCipher(aesKey)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	cipherBody := gcm.Seal(nil, nonce, expectedBody, nil)
	validBody := append(nonce, cipherBody...)
	tests := []struct {
		name         string
		rsaKey       *rsa.PrivateKey
		cryptoHeader string
		requestBody  []byte
		expStatus    int
		expBody      []byte
		check        bool
	}{
		{
			name:         "положительный тест",
			rsaKey:       privateKey,
			cryptoHeader: validHeader,
			requestBody:  validBody,
			expStatus:    http.StatusOK,
			expBody:      expectedBody,
			check:        true,
		},
		{
			name:         "ключ RSA равен nil",
			rsaKey:       nil,
			cryptoHeader: "",
			requestBody:  expectedBody,
			expStatus:    http.StatusOK,
			expBody:      expectedBody,
			check:        true,
		},
		{
			name:         "отсутствует заголовок Crypto-Key",
			rsaKey:       privateKey,
			cryptoHeader: "",
			requestBody:  expectedBody,
			expStatus:    http.StatusBadRequest,
			check:        false,
		},
		{
			name:         "невалидные данные в заголовке",
			rsaKey:       privateKey,
			cryptoHeader: "",
			requestBody:  expectedBody,
			expStatus:    http.StatusBadRequest,
			check:        false,
		},
		{
			name:         "тело запроса короче размера Nonce",
			rsaKey:       privateKey,
			cryptoHeader: validHeader,
			requestBody:  []byte("test"),
			expStatus:    http.StatusBadRequest,
			check:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := CryptoMiddleware(tt.rsaKey)
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.check {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, tt.expBody, body)
					var metrics []models.Metrics
					err = json.Unmarshal(body, &metrics)
					assert.NoError(t, err)
					assert.Equal(t, metrics, expMetrics)
				}
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.requestBody))
			if tt.cryptoHeader != "" {
				req.Header.Set("Crypto-Key", tt.cryptoHeader)
			}
			rec := httptest.NewRecorder()
			middleware(h).ServeHTTP(rec, req)
			assert.Equal(t, tt.expStatus, rec.Code)
		})
	}
}
