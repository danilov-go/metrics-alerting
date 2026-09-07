package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type hashWriter struct {
	w   http.ResponseWriter
	key string
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) WriteHeader(statusCode int) {
	h.w.WriteHeader(statusCode)
}

func (h *hashWriter) Write(b []byte) (int, error) {
	hs := hmac.New(sha256.New, []byte(h.key))
	hs.Write(b)
	hash := hex.EncodeToString(hs.Sum(nil))
	h.Header().Set("HashSHA256", hash)
	return h.w.Write(b)
}

// HashMiddleware проверяет HMAC-SHA256 подпись входящих запросов в заголовке HashSHA256
// и автоматически подписывает исходящие ответы сервера.
func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}
			expHash := r.Header.Get("HashSHA256")
			if expHash == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			expectedHash, err := hex.DecodeString(expHash)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(body))
			hs := hmac.New(sha256.New, []byte(key))
			_, _ = hs.Write(body)
			hash := hs.Sum(nil)
			if !hmac.Equal(hash, expectedHash) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			hashStr := &hashWriter{
				w:   w,
				key: key,
			}
			h.ServeHTTP(hashStr, r)
		})
	}
}
