package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

type PrivKey []byte
type PubKey []byte
type Cipher cipher.AEAD

func decodeKey(b64 string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(b64)

	if err != nil {
		return nil, err
	}
	if len(key) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size")
	}
	return key, nil
}

func DecodePrivKey(b64 string) (PrivKey, error) {
	return decodeKey(b64)
}

func DecodePubKey(b64 string) (PubKey, error) {
	return decodeKey(b64)
}

func encodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func (priv PrivKey) Encode() string {
	return encodeKey(priv)
}

func (pub PubKey) Encode() string {
	return encodeKey(pub)
}

func GeneratePrivKey() (PrivKey, error) {
	buf := make([]byte, chacha20poly1305.KeySize)

	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (priv PrivKey) PubKey() (PubKey, error) {
	pub, err := curve25519.X25519(priv, curve25519.Basepoint)

	if err != nil {
		return nil, err
	}
	return PubKey(pub), nil
}

func (priv PrivKey) Cipher(pub PubKey) (Cipher, error) {
	shared, err := curve25519.X25519(priv, pub)

	if err != nil {
		return nil, err
	}
	return chacha20poly1305.New(shared)
}

func Encrypt(cipher Cipher, plaintext []byte) ([]byte, error) {
	buf := make([]byte, cipher.NonceSize(), cipher.NonceSize()+cipher.Overhead()+len(plaintext))

	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return cipher.Seal(buf, buf, plaintext, nil), nil
}

func Decrypt(cipher Cipher, dst, ciphertext []byte) ([]byte, error) {
	nonceSize := cipher.NonceSize()

	if len(ciphertext) < nonceSize {
		return nil, errors.New("invalid ciphertext")
	}
	return cipher.Open(dst[:0], ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
}
