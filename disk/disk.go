package disk

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"soul"
	"soul/crypt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

const DefaultBucketName = "temp"

type NoteRepository struct {
	encrypter  soul.Encrypter
	decrypter  soul.Decrypter
	folderHash string
	db         *bolt.DB
	// simErr holds any error encountered during background simulation
	simErr error
}

type Note struct {
	Version int
	ID      string
	Text    string
}

func (nr *NoteRepository) Update(note *soul.Note) error {
	return nr.upsertNote(note)
}

func (nr *NoteRepository) Create(note *soul.Note) error {
	if len(strings.TrimSpace(note.ID)) != 0 {
		return fmt.Errorf("note is already assigned an id")
	}

	return nr.upsertNote(note)
}

func (nr *NoteRepository) GetAll() ([]soul.Note, error) {
	encrypted, err := nr.getRawFolder(nr.folderHash)
	if err != nil {
		return nil, err
	}

	if len(encrypted) == 0 {
		return make([]soul.Note, 0), nil
	}

	decrypted, err := nr.decrypter.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt %w", err)
	}

	// now deserialize
	// TODO: Dealloc diskformat
	diskFormat := new([]Note)
	decoder := gob.NewDecoder(bytes.NewReader(decrypted))
	err = decoder.Decode(diskFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode notes %w", err)
	}

	// convert from disk to soul
	var notes []soul.Note
	for _, note := range *diskFormat {
		notes = append(notes, soul.Note{
			ID:      note.ID,
			Version: soul.Version(note.Version),
			Text:    soul.NewBindingFromString(note.Text),
		})
	}

	return notes, nil
}

func (nr *NoteRepository) UpdateAll(notes []soul.Note) error {
	err := nr.db.Update(func(tx *bolt.Tx) error {
		return nr.saveAllTx(tx, notes)
	})

	if err != nil {
		return err
	}

	return nil
}

func (nr *NoteRepository) saveAllTx(tx *bolt.Tx, notes []soul.Note) error {
	// TODO: Dealloc disk format after use
	var diskFormat []Note
	for _, note := range notes {
		txt, err := note.Text.Get()
		if err != nil {
			return err
		}

		diskFormat = append(diskFormat, Note{
			ID:      note.ID,
			Version: int(note.Version),
			Text:    txt,
		})
	}
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(diskFormat)
	if err != nil {
		return fmt.Errorf("failed to encode %w", err)
	}

	encrypted, err := nr.encrypter.Encrypt(encoded.Bytes())
	if err != nil {
		return fmt.Errorf("failed to encrypt %w", err)
	}

	b := tx.Bucket([]byte(DefaultBucketName))
	err = b.Put([]byte(nr.folderHash), encrypted)

	if err != nil {
		return fmt.Errorf("failed to update folder %w", err)
	}

	return nil
}

func (nr *NoteRepository) getRawFolder(folder string) ([]byte, error) {
	var result []byte
	nr.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DefaultBucketName))
		fetched := b.Get([]byte(folder))
		result = make([]byte, len(fetched))
		copy(result, fetched)

		return nil
	})

	return result, nil
}

func (nr *NoteRepository) upsertNote(note *soul.Note) error {
	err := nr.db.Update(func(tx *bolt.Tx) error {
		existing, err := nr.GetAll()
		if err != nil {
			return err
		}

		if len(strings.TrimSpace(note.ID)) == 0 {
			note.ID = uuid.NewString()
			existing = append(existing, *note)
			return nr.saveAllTx(tx, existing)
		}

		// find the index
		foundIndex := -1
		for i := 0; i < len(existing); i++ {
			if existing[i].ID == note.ID {
				foundIndex = i
				break
			}
		}

		if foundIndex == -1 {
			return fmt.Errorf("note not found in disk, cannot update")
		}

		existing[foundIndex] = *note
		return nr.saveAllTx(tx, existing)
	})

	if err != nil {
		return err
	}

	return nil

}

func NewNoteRepository(dbPath, folder string, password string, encrypterFunc func(string) (soul.Encrypter, error), decrypterFunc func(string) (soul.Decrypter, error)) (*NoteRepository, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init db %w", err)
	}

	return newNoteRepository(db, folder, password, encrypterFunc, decrypterFunc, false, nil)
}

func NewNoteRepositoryWithDb(db *bolt.DB, folder string, password string, encrypterFunc func(string) (soul.Encrypter, error), decrypterFunc func(string) (soul.Decrypter, error)) (*NoteRepository, error) {
	return newNoteRepository(db, folder, password, encrypterFunc, decrypterFunc, false, nil)
}

func NewNoteRepositoryWithLoadSim(dbPath, folder string, password string, encrypterFunc func(string) (soul.Encrypter, error), decrypterFunc func(string) (soul.Decrypter, error), enableLoadSim bool, loadSimExceptions []string) (*NoteRepository, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init db %w", err)
	}

	return newNoteRepository(db, folder, password, encrypterFunc, decrypterFunc, enableLoadSim, loadSimExceptions)
}

func NewNoteRepositoryWithDbAndLoadSim(db *bolt.DB, folder string, password string, encrypterFunc func(string) (soul.Encrypter, error), decrypterFunc func(string) (soul.Decrypter, error), enableLoadSim bool, loadSimExceptions []string) (*NoteRepository, error) {
	return newNoteRepository(db, folder, password, encrypterFunc, decrypterFunc, enableLoadSim, loadSimExceptions)
}

func newNoteRepository(db *bolt.DB, folder string, password string, encrypterFunc func(string) (soul.Encrypter, error), decrypterFunc func(string) (soul.Decrypter, error), enableLoadSim bool, loadSimExceptions []string) (*NoteRepository, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DefaultBucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	passwordPartialHash, err := crypt.CalculateHash([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("unable to hash pwd %w", err)
	}

	passwordStep2, err := crypt.CalculateHash(append(passwordPartialHash, []byte(folder)...))
	if err != nil {
		return nil, fmt.Errorf("unable to hash pwd %w", err)
	}

	var finalPassword []byte
	for i := 0; i < len(passwordStep2); i = i + 2 {
		finalPassword = append(finalPassword, passwordStep2[i])
	}

	pwdByte := hex.EncodeToString(finalPassword)

	encrypter, err := encrypterFunc(pwdByte)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypter %w", err)
	}

	decrypter, err := decrypterFunc(pwdByte)
	if err != nil {
		return nil, fmt.Errorf("failed to create decrypter %w", err)
	}

	folderHash, err := crypt.CalculateStringHash(folder)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate folder name hash %w", err)
	}

	repo := &NoteRepository{
		encrypter:  encrypter,
		decrypter:  decrypter,
		db:         db,
		folderHash: folderHash,
	}

	if enableLoadSim {
		// start load simulation service
		simulator, err := NewLoadSimulator(db, loadSimExceptions, func(key string) (soul.Encrypter, error) {
			crypter, err := crypt.NewCryptor(key)
			if err != nil {
				return nil, err
			}

			return crypter, nil
		}, func(err error) {
			repo.simErr = err
		})

		if err != nil {
			return nil, err
		}

		simulator.Start()
	}

	return repo, nil
}
