package crypt_test

import (
	"encoding/hex"
	"soul/crypt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const key = "passphrasel"

func TestCryptorEncryptDecrypt(t *testing.T) {
	t.Parallel()

	var input = []byte("wow this is amazing")
	cryptor, err := crypt.NewCryptor(key)
	assert.Nil(t, err)
	encrypted, err := cryptor.Encrypt(input)
	assert.Nil(t, err)

	cryptor2, err := crypt.NewCryptor(key)
	assert.Nil(t, err)
	original, err := cryptor2.Decrypt(encrypted)
	assert.Nil(t, err)
	assert.Equal(t, input, original)

	encrypted2, err := cryptor2.Encrypt(input)
	assert.Nil(t, err)

	original2, err := cryptor.Decrypt(encrypted2)
	assert.Nil(t, err)
	assert.Equal(t, input, original2)
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	var input = []byte("wow this is amazing")
	encoded := "e1e722879bca0d0fa9a4f6f803527e2bd74e8394a7b19816e409af69f9bd44594c5251b9dc0445634aea1a1c1a64ac"

	decoded, _ := hex.DecodeString(encoded)
	cryptor2, err := crypt.NewCryptor(key)
	assert.Nil(t, err)
	original, err := cryptor2.Decrypt(decoded)
	assert.Nil(t, err)
	assert.Equal(t, input, original)
}
