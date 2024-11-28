package fhcore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/feihan-im/openapi-sdk-go/internal/model"
)

func encryptMessage(secret string, data []byte) (*model.SecureMessage, error) {
	timestamp := uint64(time.Now().UnixMilli())
	nonce := randomAlphaNumString(16)
	initKey := sha256.Sum256([]byte(fmt.Sprintf(
		"%d:%s:%s",
		timestamp,
		secret,
		nonce,
	)))
	key := randomBytes(256 / 8)
	encryptedKey, err := encryptAES256CBC(key, initKey[:])
	if err != nil {
		return nil, err
	}
	encryptedData, err := encryptAES256CBC(data, key)
	if err != nil {
		return nil, err
	}
	return &model.SecureMessage{
		Version:       defaultSecureVersion,
		Timestamp:     timestamp,
		Nonce:         nonce,
		EncryptedKey:  encryptedKey,
		EncryptedData: encryptedData,
	}, nil
}

func decryptMessage(secret string, message *model.SecureMessage) ([]byte, error) {
	switch message.Version {
	case defaultSecureVersion:
		{
			initKey := sha256.Sum256([]byte(fmt.Sprintf(
				"%d:%s:%s",
				message.Timestamp,
				secret,
				message.Nonce,
			)))
			key, err := decryptAES256CBC(message.EncryptedKey, initKey[:])
			if err != nil {
				return nil, err
			}
			data, err := decryptAES256CBC(message.EncryptedData, key)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("unsupport version %s", message.Version)
}

func encryptAES256CBC(data []byte, key []byte) ([]byte, error) {
	data = pkcs7Pad(data, aes.BlockSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, aes.BlockSize+len(data))
	iv := out[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(out[aes.BlockSize:], data)

	return out, nil
}

func decryptAES256CBC(data []byte, key []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}
	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("data is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	out := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(out, data)

	return pkcs7Unpad(out), nil
}

func pkcs7Pad(b []byte, blocksize int) []byte {
	n := blocksize - (len(b) % blocksize)
	return append(b, bytes.Repeat([]byte{byte(n)}, n)...)
}

func pkcs7Unpad(b []byte) []byte {
	c := b[len(b)-1]
	n := int(c)
	return b[:len(b)-n]
}

const alphaNumLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomBytes(size int) []byte {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return b
}

func randomAlphaNumString(size int) string {
	b := randomBytes(size)
	for i := range b {
		b[i] = alphaNumLetters[int(b[i])%len(alphaNumLetters)]
	}
	return string(b)
}
