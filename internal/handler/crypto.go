package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
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
			aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, rsaKey, nil)
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
			block, err := aes.NewCipher(aesKey)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			gcm, err := cipher.NewGCM(block)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			nonceSize := gcm.NonceSize()
			if len(body) < nonceSize {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			nonce := body[:nonceSize]
			cipherBody := body[nonceSize:]
			decryptedBody, err := gcm.Open(nil, nonce, cipherBody, nil)
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
