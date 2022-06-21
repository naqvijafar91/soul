//go:build load
// +build load

package disk_test

import (
	"fmt"
	"os"
	"soul"
	"soul/crypt"
	"soul/disk"
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// The tests in this file is used to create a big database and then run the application to understand real world app behaviour.

func TestLoad(t *testing.T) {
	t.Parallel()

	var dbPath = os.Getenv(soul.DBPathKeyName)
	assert.NotZero(t, len(strings.TrimSpace(dbPath)))

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	old, err := disk.GetKeysCount(db)
	assert.Nil(t, err)

	fmt.Println(db.GoString())
	fmt.Println(db.String())
	folderCount := 1000

	for i := 0; i < folderCount; i++ {
		now := time.Now()

		folderName := uuid.NewString()
		key := uuid.NewString()

		repo, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, folderName, key[:15], crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, false, nil)
		assert.Nil(t, err)

		notes := disk.GenerateRandomNotes()
		fmt.Println(fmt.Sprintf("time taken to generate %d notes is %v", len(notes), time.Since(now)))
		assert.Nil(t, repo.UpdateAll(notes))
		fetched, err := repo.GetAll()
		assert.Nil(t, err)

		assert.Equal(t, notes, fetched)

		fmt.Println("total time taken to execute 1 iteration is ", time.Since(now))
	}

	updated, err := disk.GetKeysCount(db)
	assert.Nil(t, err)
	assert.Equal(t, old+uint64(folderCount), updated)
}
