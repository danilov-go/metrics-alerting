package handler_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	body := "test gzip"
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(body))
		assert.NoError(t, err)
	})
	h := handler.GzipMiddleware(next)
	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		wantCode        int
		wantErr         bool
	}{
		{
			name:           "без сжатия ",
			acceptEncoding: "",
			wantCode:       http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "со сжатием",
			acceptEncoding: "gzip",
			wantCode:       http.StatusOK,
			wantErr:        true,
		},
		{
			name:            "некорректный gzip",
			acceptEncoding:  "",
			contentEncoding: "gzip",
			wantCode:        http.StatusInternalServerError,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test-gzip", bytes.NewReader([]byte("plain text")))
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			assert.Equal(t, tt.wantCode, res.Code)
			if res.Code == http.StatusOK {
				if tt.wantErr {
					assert.True(t, strings.Contains(res.Header().Get("Content-Encoding"), "gzip"))
					zr, err := gzip.NewReader(res.Body)
					require.NoError(t, err)
					defer func() { _ = zr.Close() }()
					uncompressedResult, err := io.ReadAll(zr)
					require.NoError(t, err)
					assert.Equal(t, body, string(uncompressedResult))
				} else {
					assert.NotContains(t, res.Header().Get("Content-Encoding"), "gzip")
					assert.Equal(t, body, res.Body.String())
				}
			}
		})
	}
}
