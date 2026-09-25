package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestHTTPSender(t *testing.T) {
	logger := zaptest.NewLogger(t)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	value := 123.45
	delta := int64(5)
	metrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	}
	expJSON, err := json.Marshal(metrics)
	require.NoError(t, err)
	testKey := "secret-key"
	testIP := "localhost:8080"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, testIP, r.Header.Get("X-Real-IP"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		cryptoKeyHex := r.Header.Get("Crypto-Key")
		if cryptoKeyHex != "" {
			cipherKey, err := hex.DecodeString(cryptoKeyHex)
			require.NoError(t, err)
			body, err = crypto.DecryptBody(cipherKey, body, privateKey)
			require.NoError(t, err)
		}
		gzipReader, err := gzip.NewReader(bytes.NewReader(body))
		require.NoError(t, err)
		defer gzipReader.Close()
		unzippedJSON, err := io.ReadAll(gzipReader)
		require.NoError(t, err)
		assert.JSONEq(t, string(expJSON), string(unzippedJSON))
		expHash := r.Header.Get("HashSHA256")
		if expHash != "" {
			hs := hmac.New(sha256.New, []byte(testKey))
			hs.Write(unzippedJSON)
			expectedHash := hex.EncodeToString(hs.Sum(nil))
			assert.Equal(t, expectedHash, expHash)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	sender := NewHTTPSender(u.Host, testKey, testIP, logger.Sugar())
	sender.Send(context.Background(), metrics, publicKey)
}
