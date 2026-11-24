// Package crypto seals small secrets (API keys) stored in the database with AES-256-GCM.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// KeySize is the AES-256 key length in bytes.
const KeySize = 32

// ErrDecrypt is returned for a wrong key or a tampered ciphertext.
var ErrDecrypt = errors.New("crypto: cannot decrypt")

// Sealer encrypts and decrypts with one key. Ciphertexts are base64(nonce || sealed).
type Sealer struct {
	aead cipher.AEAD
}

// ParseKey decodes a base64 (standard or URL, padded or not) 32-byte key.
func ParseKey(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if key, err := enc.DecodeString(s); err == nil {
			if len(key) != KeySize {
				return nil, fmt.Errorf("crypto: key must be %d bytes, got %d", KeySize, len(key))
			}
			return key, nil
		}
	}
	return nil, errors.New("crypto: key must be base64")
}

// NewSealer returns a Sealer for a 32-byte key.
func NewSealer(key []byte) (*Sealer, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("crypto: key must be %d bytes", KeySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Sealer{aead: aead}, nil
}

// Seal encrypts plaintext with a random nonce. additionalData binds the ciphertext to its context
// (e.g. the app config code) so it can't be moved to another row.
func (s *Sealer) Seal(plaintext, additionalData []byte) (string, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := s.aead.Seal(nonce, nonce, plaintext, additionalData)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a value made by Seal with the same additionalData.
func (s *Sealer) Open(ciphertext string, additionalData []byte) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil || len(raw) < s.aead.NonceSize() {
		return nil, ErrDecrypt
	}
	nonce, sealed := raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():]
	plain, err := s.aead.Open(nil, nonce, sealed, additionalData)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plain, nil
}
