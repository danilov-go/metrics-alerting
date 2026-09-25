package crypto_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	tests := []struct {
		name      string
		publicKey *rsa.PublicKey
		body      []byte
	}{
		{
			name:      "положительный тест",
			publicKey: publicKey,
			body:      []byte(""),
		},
		{
			name:      "положительный тест",
			publicKey: publicKey,
			body:      []byte(`[{"id":"Alloc","type":"gauge","value":123.45}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherBody, cipherKey, err := crypto.EncryptBody(tt.publicKey, tt.body)
			assert.NoError(t, err)
			assert.NotEmpty(t, cipherBody)
			assert.NotEmpty(t, cipherKey)
			aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, cipherKey, nil)
			require.NoError(t, err)
			assert.Len(t, aesKey, 32)
			block, err := aes.NewCipher(aesKey)
			require.NoError(t, err)
			gsm, err := cipher.NewGCM(block)
			require.NoError(t, err)
			nonceSize := gsm.NonceSize()
			require.GreaterOrEqual(t, len(cipherBody), nonceSize)
			nonce := cipherBody[:nonceSize]
			body := cipherBody[nonceSize:]
			decryptedBody, err := gsm.Open(nil, nonce, body, nil)
			require.NoError(t, err)
			assert.Equal(t, string(tt.body), string(decryptedBody))
			finalBody, err := crypto.DecryptBody(cipherKey, cipherBody, privateKey)
			assert.NoError(t, err)
			assert.Equal(t, string(tt.body), string(finalBody))
		})
	}
}
