package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
)

func EncryptBody(publicKey *rsa.PublicKey, body []byte) ([]byte, []byte, error) {
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

func DecryptBody(rsaKey []byte, body []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, rsaKey, nil)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(body) < nonceSize {
		return nil, io.ErrUnexpectedEOF
	}
	nonce := body[:nonceSize]
	cipherBody := body[nonceSize:]
	decryptedBody, err := gcm.Open(nil, nonce, cipherBody, nil)
	if err != nil {
		return nil, err
	}
	return decryptedBody, nil
}
