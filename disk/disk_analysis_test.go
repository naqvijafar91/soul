//go:build analysis
// +build analysis

package disk_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"soul"
	"soul/crypt"
	"soul/disk"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

//NOTE: This file contains tests for analyzing boltdb behaviour with respect to data changes, these tests do not need to run, I have kept
// them to track boltdb changes in future versions, or maybe if i replace with other key val stores.

func TestFileSignatureLoadSimulation(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", "test")

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	repo, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, "temp-0", "dummmy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, true, []string{"exception"})
	assert.Nil(t, err)

	notes, err := repo.GetAll()
	assert.Nil(t, err)
	assert.NotEmpty(t, notes)

	// calculate file signature array
	original := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, original)

	time.Sleep(20 * time.Second)

	after20Sec := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, after20Sec)
	assert.NotEqual(t, original, after20Sec)

	// get sig difference
	modifiedNew, addedNew := calculateSigDifference(t, original, after20Sec)
	assert.NotEmpty(t, modifiedNew)
	t.Logf("modifieed %v(percent), added %v(percent)", modifiedNew, addedNew)
	assert.GreaterOrEqual(t, modifiedNew, float64(8))
}

func TestFileSignatureWithoutLoadSimulation(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", "test")

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	repo, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, "temp-0", "dummy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, false, nil)
	assert.Nil(t, err)

	notes, err := repo.GetAll()
	assert.Nil(t, err)
	assert.NotEmpty(t, notes)

	// calculate file signature array
	original := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, original)

	time.Sleep(20 * time.Second)

	after20Sec := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, after20Sec)
	assert.Equal(t, original, after20Sec)
}

func TestFileSignatureAll(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", "test")

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	repo, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, "temp-0", "dummy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, false, nil)
	assert.Nil(t, err)

	notes, err := repo.GetAll()
	assert.Nil(t, err)
	assert.NotEmpty(t, notes)

	// calculate file signature array
	originalSig := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, originalSig)

	// insert 1 value and then calculate signature
	note := &notes[0]
	note.Text.Set(fmt.Sprintf("%s %s ", disk.Lorel, uuid.NewString()))
	assert.Nil(t, repo.Update(note))

	afterSingleUpdate := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, afterSingleUpdate)
	assert.NotEqual(t, originalSig, afterSingleUpdate)

	// get sig difference
	modified, added := calculateSigDifference(t, originalSig, afterSingleUpdate)
	assert.NotEmpty(t, modified)
	fmt.Println(added)

	countUpdateItems := 0
	// .. now update all keys with same value and then see
	assert.Nil(t, db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(disk.DefaultBucketName))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			countUpdateItems++
			assert.Nil(t, b.Put(k, v))

		}

		return nil
	}))

	afterAllUpdate := calculateFileChunkSignature(t, dbPath)
	assert.NotEmpty(t, afterSingleUpdate)
	assert.NotEqual(t, originalSig, afterSingleUpdate)

	// get sig difference
	modifiedNew, addedNew := calculateSigDifference(t, afterSingleUpdate, afterAllUpdate)
	assert.NotEmpty(t, modifiedNew)
	assert.GreaterOrEqual(t, modifiedNew+addedNew, float64(20.0))
}

// calculateSigDifference calculates sig difference
func calculateSigDifference(t *testing.T, old, new []string) (modifiedPercentage, addedPercentage float64) {
	added := 0.0
	modified := 0.0
	for i := 0; i < len(old) && i < len(new); i++ {
		if new[i] != old[i] {
			modified++
		}
	}

	added = float64(len(new) - len(old))

	modifiedPercentage = ((modified) / float64(len(old))) * 100
	addedPercentage = (added / float64(len(old))) * 100

	return
}

func TestDataComparison(t *testing.T) {
	t.Parallel()

	var dbPath = fmt.Sprintf("./tmp/%s.db", uuid.NewString())

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	cryptor, err := crypt.NewCryptor("dummmy key")
	assert.Nil(t, err)

	folder := "temp"
	repo1, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, folder, "dummy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, false, nil)
	assert.Nil(t, err)
	insertData(t, folder, repo1)

	assert.Nil(t, db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(disk.DefaultBucketName))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			newDec := make([]byte, len(v))
			copy(newDec, v)
			newDec = append(newDec, []byte(disk.Lorel)[:5]...)
			encrypted, err := cryptor.Encrypt(newDec)
			if err != nil {
				t.Fatalf("failed to encrypt %v", err)
			}

			assert.Nil(t, b.Put(k, encrypted))
			readEncerypted := b.Get(k)

			hexOrig := hex.EncodeToString(v)
			hexNew := hex.EncodeToString(readEncerypted)
			hexNotWritten := hex.EncodeToString(encrypted)

			WriteFile(t, hexOrig, "./tmp/orig.txt")

			WriteFile(t, hexNew, "./tmp/new.txt")
			WriteFile(t, hexNotWritten, "./tmp/not-written.txt")
		}

		return nil
	}))
}

func WriteFile(t *testing.T, data string, path string) {
	t.Helper()

	assert.Nil(t, os.WriteFile(path, []byte(data), 0644))
}

func insertData(t *testing.T, folderName string, repo *disk.NoteRepository) {
	t.Helper()

	notes, err := repo.GetAll()
	assert.Nil(t, err)
	assert.Empty(t, notes)

	assert.Nil(t, repo.Update(&soul.Note{
		Text: soul.NewBindingFromString(fmt.Sprintf("%s - %s", disk.Lorel, uuid.NewString())),
	}))

	updated, err := repo.GetAll()
	assert.Nil(t, err)
	assert.NotEmpty(t, updated)
}

func seedData(t *testing.T, cryptor *crypt.Crypter) {
	t.Helper()

	var dbPath = fmt.Sprintf("./tmp/%s.db", "test")

	cryptor, err := crypt.NewCryptor("dummmy key")
	assert.Nil(t, err)

	db, err := bolt.Open(dbPath, 0600, nil)
	assert.Nil(t, err)

	// create 100 folders with 100 notes each
	for i := 0; i < 100; i++ {
		repo, err := disk.NewNoteRepositoryWithDbAndLoadSim(db, fmt.Sprintf("%s-%d", "temp", i), "dummy key", crypt.NewSoulEncrypter, crypt.NewSoulDecrypter, false, nil)
		assert.Nil(t, err)

		// create 10 notes
		const times = 10
		for i := 0; i < times; i++ {
			assert.Nil(t, repo.Update(&soul.Note{
				Text: soul.NewBindingFromString(fmt.Sprintf("%s - %d", disk.Lorel, i)),
			}))
		}
	}
}

func calculateFileChunkSignature(t *testing.T, filePath string) []string {
	t.Helper()

	var result []string

	f, err := os.Open(filePath)
	if err != nil {
		assert.Nil(t, err)
	}
	defer f.Close()

	const BufferSize = 4 * 1024
	buffer := make([]byte, BufferSize)

	for {
		_, err := f.Read(buffer)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}

			break
		}
		h := sha256.New()
		if _, err := io.Copy(h, bytes.NewReader(buffer)); err != nil {
			assert.Nil(t, err)
		}

		result = append(result, hex.EncodeToString(h.Sum(nil)))
	}

	return result
}

func calculateFileHash(t *testing.T, filePath string) string {
	t.Helper()

	f, err := os.Open(filePath)
	if err != nil {
		assert.Nil(t, err)
	}

	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		assert.Nil(t, err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
