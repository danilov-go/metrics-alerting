package handler

import (
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"
)

// CryptoMiddleware расшифровывает тело запроса.
func CryptoMiddleware(key *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == nil {
				next.ServeHTTP(w, r)
				return
			}
			cryptoKey := r.Header.Get("Crypto-Key")
			if cryptoKey == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			rsaKey, err := hex.DecodeString(cryptoKey)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			defer r.Body.Close()
			decryptedBody, err := crypto.DecryptBody(rsaKey, body, key)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))
			next.ServeHTTP(w, r)
		})
	}
}
