package soul

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ReadNotes reades notes from the source
type ReadNotes func() ([]Note, error)

// UpdateNote updates the note
type UpdateNote func(note *Note) error

// WriteErr writes an error
type WriteErr func(error)

type SyncService struct {
	noteIndex     map[string]string
	readNotesFunc ReadNotes
	updateNote    UpdateNote
	interval      time.Duration
	writeErr      WriteErr
	quitChan      chan bool
	running       bool
}

func (ss *SyncService) Start() {
	if ss.running {
		return
	}

	ticker := time.NewTicker(ss.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := ss.executeOnce()
				if err != nil {
					ss.writeErr(err)
				}
			case <-ss.quitChan:
				ticker.Stop()
				ss.running = false
				return
			}
		}
	}()

	ss.running = true
}

func (ss *SyncService) Stop() {
	ss.quitChan <- true
}

func (ss *SyncService) executeOnce() error {
	notes, err := ss.readNotesFunc()
	if err != nil {
		return fmt.Errorf("unable to retrieve notes %w", err)
	}

	for _, note := range notes {
		currentSignature, err := calculateSignature(&note)
		if err != nil {
			return err
		}

		if ss.noteIndex[note.ID] != currentSignature {
			err = ss.updateNote(&note)
			if err != nil {
				return fmt.Errorf("failed to update note %s %w", note.ID, err)
			}

			ss.noteIndex[note.ID] = currentSignature
		}
	}

	return nil
}

func calculateSignature(note *Note) (string, error) {
	txt, err := note.Text.Get()
	if err != nil {
		return "", fmt.Errorf("unable to read text from note %w", err)
	}

	hashFunc := sha256.New()
	_, err = hashFunc.Write([]byte(txt))
	if err != nil {
		return "", fmt.Errorf("unable to hash the note for ID %s because %w", note.ID, err)
	}

	return hex.EncodeToString(hashFunc.Sum(nil)), nil
}

// NewSyncService creates a new sync service that can be started and stopped
func NewSyncService(readNotesFunc ReadNotes, updateNote UpdateNote, writeErr WriteErr,
	interval time.Duration) *SyncService {
	index := make(map[string]string)
	initialNotes, _ := readNotesFunc()
	for _, note := range initialNotes {
		signature, _ := calculateSignature(&note)
		index[note.ID] = signature
	}

	return &SyncService{
		noteIndex:     index,
		quitChan:      make(chan bool),
		interval:      interval,
		readNotesFunc: readNotesFunc,
		updateNote:    updateNote,
		writeErr:      writeErr,
	}
}
