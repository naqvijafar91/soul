package soul

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"
)

// Encrypter encrypts the data
type Encrypter interface {
	Encrypt(input []byte) ([]byte, error)
}

// Decrypter decrypts the data
type Decrypter interface {
	Decrypt([]byte) ([]byte, error)
}

// ConfigStore stores configurations
type ConfigStore interface {
	SetString(string, string)
	GetString(string) string
	Delete(string)
}

// Credentials represents a user credential
type Credentials struct {
	Identifier string
	Password   string
}

const LocalCreditialsKeyName = "ENCRYPTED_DATA_MAIN"

const DBPathKeyName = "DB_PATH"

func StoreDbPath(store ConfigStore, path string) {
	store.SetString(DBPathKeyName, path)
}

func GetDBPath(store ConfigStore) string {
	return store.GetString(DBPathKeyName)
}

// IsSignedIn checks if the user is signed in locally
func IsSignedIn(store ConfigStore) bool {
	encrypted := store.GetString(LocalCreditialsKeyName)
	return strings.TrimSpace(encrypted) != ""
}

func GetCredentials(store ConfigStore, decrypter Decrypter) (*Credentials, error) {
	serializedData := store.GetString(LocalCreditialsKeyName)
	if strings.TrimSpace(serializedData) == "" {
		return nil, fmt.Errorf("credentials not found")
	}

	encrypted, err := base64.StdEncoding.DecodeString(serializedData)
	if err != nil {
		return nil, fmt.Errorf("failed to base 64 decode serialized data")
	}

	decrypted, err := decrypter.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt %w", err)
	}

	credentials := new(Credentials)
	decoder := gob.NewDecoder(bytes.NewReader(decrypted))
	err = decoder.Decode(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credentials %w", err)
	}

	if credentials == nil {
		return nil, fmt.Errorf("corrupted credentials")
	}

	return credentials, nil
}

func SetCredentials(store ConfigStore, encrypter Encrypter, credentials *Credentials) error {
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(*credentials)
	if err != nil {
		return fmt.Errorf("failed to encode %w", err)
	}

	encrypted, err := encrypter.Encrypt(encoded.Bytes())
	if err != nil {
		return fmt.Errorf("failed to encrypt %w", err)
	}

	store.SetString(LocalCreditialsKeyName, base64.StdEncoding.EncodeToString(encrypted))

	return nil
}

func DeleteCredentials(store ConfigStore) {
	store.Delete(LocalCreditialsKeyName)
}
