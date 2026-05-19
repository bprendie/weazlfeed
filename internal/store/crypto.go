package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"
)

const encryptedPrefix = "wfenc:v1:"

func (s *Store) encryptText(value string) (string, error) {
	if value == "" || strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	gcm, err := s.gcm()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	return encryptedPrefix + base64.RawStdEncoding.EncodeToString(nonce) + ":" +
		base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (s *Store) decryptText(value string) string {
	if !isEncryptedText(value) {
		return value
	}
	parts := strings.Split(value, ":")
	if len(parts) != 4 {
		return value
	}
	nonce, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return value
	}
	ciphertext, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return value
	}
	gcm, err := s.gcm()
	if err != nil {
		return value
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return value
	}
	return string(plain)
}

func isEncryptedText(value string) bool {
	return strings.HasPrefix(value, encryptedPrefix)
}

func (s *Store) feedKey(url string) string {
	url = s.decryptText(url)
	mac := hmac.New(sha256.New, s.key)
	_, _ = mac.Write([]byte(strings.TrimSpace(url)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Store) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
