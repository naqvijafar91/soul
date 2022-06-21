package soul_test

import (
	"fmt"
	"soul"
	"soul/crypt"
	"testing"

	fyneapp "fyne.io/fyne/v2/app"
	"github.com/stretchr/testify/assert"
)

func TestSetGetCredentials(t *testing.T) {
	t.Parallel()

	testCreds := &soul.Credentials{
		Identifier: "dummy@jsjsjsj2222jj2",
		Password:   "dummyKey@567",
	}

	// write data in a private scope
	{
		app := fyneapp.NewWithID("org.testing.soul")
		app.Preferences().RemoveValue(soul.LocalCreditialsKeyName)
		cryptor, err := crypt.NewCryptor(testCreds.Password)
		assert.Nil(t, err)

		cred, err := soul.GetCredentials(app, cryptor)
		assert.Nil(t, cred)
		assert.Equal(t, fmt.Errorf("credentials not found"), err)

		// Set once and retrieve multiple times
		assert.Nil(t, soul.SetCredentials(app, cryptor, testCreds))
	}

	for i := 0; i < 10; i++ {
		app := fyneapp.NewWithID("org.testing.soul")
		newCryptor, err := crypt.NewCryptor(testCreds.Password)
		assert.Nil(t, err)
		fetchedCreds, err := soul.GetCredentials(app, newCryptor)
		assert.Nil(t, err)
		assert.Equal(t, testCreds, fetchedCreds)
	}
}
