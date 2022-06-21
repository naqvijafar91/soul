package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"soul"
)

type Crypter struct {
	Key []byte
}

// filler is used to append keys in case if they are less than 32 bytes
// WARNING: Do not modify this, else the entire encryption decryption might fail if its old
const filler = byte('u')

var _ soul.Encrypter = &Crypter{}
var _ soul.Decrypter = &Crypter{}

func (crypter *Crypter) Encrypt(input []byte) ([]byte, error) {
	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(crypter.Key)
	// if there are any errors, handle them
	if err != nil {
		return nil, fmt.Errorf("failed to create cypher %w", err)
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	// if any error generating new GCM
	// handle them
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm cypher %w", err)
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to populate nonce %w", err)
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, input, nil), nil
}

func (crypter *Crypter) Decrypt(encrypted []byte) ([]byte, error) {
	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(crypter.Key)
	// if there are any errors, handle them
	if err != nil {
		return nil, fmt.Errorf("failed to create cypher %w", err)
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm cypher %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("failed to populate nonce %w", err)
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt %w", err)
	}

	return plaintext, nil
}

// TODO: GCM vs XORKeyStream

// Encrypt method is to encrypt or hide any classified text
// func Encrypt(text, MySecret string) (string, error) {
// 	block, err := aes.NewCipher([]byte(MySecret))
// 	if err != nil {
// 		return "", err
// 	}
// 	plainText := []byte(text)
// 	cfb := cipher.NewCFBEncrypter(block, bytes)
// 	cipherText := make([]byte, len(plainText))
// 	cfb.XORKeyStream(cipherText, plainText)
// 	return Encode(cipherText), nil
// }

func CalculateStringHash(txt string) (string, error) {
	hash, err := CalculateHash([]byte(txt))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash), nil
}

func CalculateHash(val []byte) ([]byte, error) {
	hashFunc := sha256.New()
	_, err := hashFunc.Write(val)
	if err != nil {
		return nil, fmt.Errorf("unable to hash the folder name %w", err)
	}

	return hashFunc.Sum(nil), nil
}

func NewCryptor(keyStr string) (*Crypter, error) {
	key := []byte(keyStr)
	if len(key) > 32 {
		return nil, fmt.Errorf("key cannot be more than 32 bytes, got %d length", len(key))
	}

	for i := len(key); i < 32; i++ {
		key = append(key, filler)
	}

	return &Crypter{Key: key}, nil
}

func NewSoulEncrypter(keyStr string) (soul.Encrypter, error) {
	return NewCryptor(keyStr)
}

func NewSoulDecrypter(keyStr string) (soul.Decrypter, error) {
	return NewCryptor(keyStr)
}
