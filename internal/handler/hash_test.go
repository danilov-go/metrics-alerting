package handler_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/stretchr/testify/assert"
)

func generateHash(body []byte, key string) string {
	hs := hmac.New(sha256.New, []byte(key))
	hs.Write(body)
	return hex.EncodeToString(hs.Sum(nil))
}

func TestHashMiddleware(t *testing.T) {
	secret := "secret_key"
	testBody := []byte("test_body")
	validHash := generateHash(testBody, secret)
	type want struct {
		statusCode int
		check      bool
		respBody   string
	}
	tests := []struct {
		name        string
		key         string
		headerHash  string
		requestBody []byte
		want        want
	}{
		{
			name:        "положительный тест",
			key:         secret,
			headerHash:  validHash,
			requestBody: testBody,
			want: want{
				statusCode: http.StatusOK,
				check:      true,
				respBody:   "test",
			},
		},
		{
			name:        "пустой ключ",
			key:         "",
			headerHash:  "",
			requestBody: testBody,
			want: want{
				statusCode: http.StatusOK,
				check:      false,
				respBody:   "test",
			},
		},
		{
			name:        "заголовок HashSHA256 пустой",
			key:         secret,
			headerHash:  "",
			requestBody: testBody,
			want: want{
				statusCode: http.StatusBadRequest,
				check:      false,
			},
		},
		{
			name:        "невалидный hex в заголовке",
			key:         secret,
			headerHash:  "unknow",
			requestBody: testBody,
			want: want{
				statusCode: http.StatusBadRequest,
				check:      false,
			},
		},
		{
			name:        "подпись не совпадает",
			key:         secret,
			headerHash:  generateHash([]byte("unknow"), secret),
			requestBody: testBody,
			want: want{
				statusCode: http.StatusBadRequest,
				check:      false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("test"))
				assert.NoError(t, err)
			})
			middleware := handler.HashMiddleware(tt.key)
			handlerToTest := middleware(nextHandler)
			req, err := http.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.requestBody))
			assert.NoError(t, err)
			if tt.headerHash != "" {
				req.Header.Set("HashSHA256", tt.headerHash)
			}
			rec := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rec, req)
			assert.Equal(t, tt.want.statusCode, rec.Code)
			if tt.want.statusCode == http.StatusOK {
				assert.Contains(t, rec.Body.String(), tt.want.respBody)
				if tt.want.check {
					respHash := rec.Header().Get("HashSHA256")
					assert.NotEmpty(t, respHash)
					expectedRespHash := generateHash(rec.Body.Bytes(), tt.key)
					assert.Equal(t, expectedRespHash, respHash)
				}
			}
		})
	}
}
