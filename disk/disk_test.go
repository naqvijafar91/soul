package disk_test

import (
	"fmt"
	"soul"
	"soul/crypt"
	"soul/disk"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReadWrite(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", uuid.NewString())

	repo, err := disk.NewNoteRepository(dbPath, "temp", "dummy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter)
	assert.Nil(t, err)

	notes, err := repo.GetAll()
	assert.Nil(t, err)
	assert.Empty(t, notes)

	assert.Nil(t, repo.Update(&soul.Note{
		Text: soul.NewBindingFromString(fmt.Sprintf("%s - %d", disk.Lorel, 100)),
	}))

	updated, err := repo.GetAll()
	assert.Nil(t, err)
	assert.NotEmpty(t, updated)
}

func TestShouldGenerateDiffPassword(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", uuid.NewString())
	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	basePwd := "dummy key"
	folder1 := "folder1"
	folder1Hash, err := crypt.CalculateStringHash(folder1)
	folder1Pwd := "57e968c50cc3952c37be85391e6f1c3a"
	assert.Nil(t, err)

	folder2 := "folder2"
	folder2Hash, err := crypt.CalculateStringHash(folder2)
	folder2Pwd := "327e46efcad62b823699321932e4e09b"
	assert.Nil(t, err)

	repo1, err := disk.NewNoteRepositoryWithDb(db, folder1, basePwd, crypt.NewSoulEncrypter, crypt.NewSoulDecrypter)
	assert.Nil(t, err)
	repo2, err := disk.NewNoteRepositoryWithDb(db, folder2, basePwd, crypt.NewSoulEncrypter, crypt.NewSoulDecrypter)
	assert.Nil(t, err)

	assert.Nil(t, repo1.Update(&soul.Note{
		Text: soul.NewBindingFromString(fmt.Sprintf("%s - %d", disk.Lorel, 100)),
	}))
	assert.Nil(t, repo2.Update(&soul.Note{
		Text: soul.NewBindingFromString(fmt.Sprintf("%s - %d", disk.Lorel, 100)),
	}))

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(disk.DefaultBucketName))
		count := 0
		b.ForEach(func(k, v []byte) error {
			fmt.Println(k, v)
			count++

			keyStr := string(k)

			switch keyStr {
			case folder1Hash:
				fetched := b.Get([]byte(folder1Hash))
				decrypter, _ := crypt.NewSoulDecrypter(folder1Pwd)
				_, err := decrypter.Decrypt(fetched)
				assert.Nil(t, err)

				//ensure vise versa is not possible
				decrypter2, _ := crypt.NewSoulDecrypter(folder2Pwd)
				_, err = decrypter2.Decrypt(fetched)
				assert.NotNil(t, err)

			case folder2Hash:
				fetched := b.Get([]byte(folder2Hash))
				decrypter, _ := crypt.NewSoulDecrypter(folder2Pwd)
				_, err := decrypter.Decrypt(fetched)
				assert.Nil(t, err)

				//ensure vise versa is not possible
				decrypter2, _ := crypt.NewSoulDecrypter(folder1Pwd)
				_, err = decrypter2.Decrypt(fetched)
				assert.NotNil(t, err)
			default:
				t.Fatal("other found")
			}

			return nil
		})

		assert.Equal(t, 2, count)
		return nil
	})
}
