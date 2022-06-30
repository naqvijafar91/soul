package testhelpers

import (
	"fmt"
	"soul"
	"soul/crypt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteConfigStoreTests(t *testing.T, configStore soul.ConfigStore) {
	t.Helper()

	testCreds := &soul.Credentials{
		Identifier: "dummy@jsjsjsj2222jj2",
		Password:   "dummyKey@567",
	}

	// write data in a private scope
	{
		soul.DeleteCredentials(configStore)
		cryptor, err := crypt.NewCryptor(testCreds.Password)
		assert.Nil(t, err)

		cred, err := soul.GetCredentials(configStore, cryptor)
		assert.Nil(t, cred)
		assert.Equal(t, fmt.Errorf("credentials not found"), err)

		// Set once and retrieve multiple times
		assert.Nil(t, soul.SetCredentials(configStore, cryptor, testCreds))
	}

	for i := 0; i < 10; i++ {
		newCryptor, err := crypt.NewCryptor(testCreds.Password)
		assert.Nil(t, err)
		fetchedCreds, err := soul.GetCredentials(configStore, newCryptor)
		assert.Nil(t, err)
		assert.Equal(t, testCreds, fetchedCreds)
	}
}
