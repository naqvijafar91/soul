package disk

import (
	"bytes"
	"fmt"
	"math/rand"
	"soul"
	"soul/crypt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// LoadSimulator is an experimental servie which simulates the load on the disk database so that the application can be tested under heavy scenarios
// This service should be turned off by default
type LoadSimulator struct {
	db            *bolt.DB
	reportError   func(error)
	exceptions    []string
	encryptorFunc func(key string) (soul.Encrypter, error)
}

const Lorel = `
What is Lorem Ipsum?

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.
Why do we use it?

It is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using 'Content here, content here', making it look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for 'lorem ipsum' will uncover many web sites still in their infancy. Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).

Where does it come from?

Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC, making it over 2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney College in Virginia, looked up one of the more obscure Latin words, consectetur, from a Lorem Ipsum passage, and going through the cites of the word in classical literature, discovered the undoubtable source. Lorem Ipsum comes from sections 1.10.32 and 1.10.33 of "de Finibus Bonorum et Malorum" (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a treatise on the theory of ethics, very popular during the Renaissance. The first line of Lorem Ipsum, "Lorem ipsum dolor sit amet..", comes from a line in section 1.10.32.

The standard chunk of Lorem Ipsum used since the 1500s is reproduced below for those interested. Sections 1.10.32 and 1.10.33 from "de Finibus Bonorum et Malorum" by Cicero are also reproduced in their exact original form, accompanied by English versions from the 1914 translation by H. Rackham.
Where can I get some?

There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc.`

func (ls *LoadSimulator) Start() {
	go func() {
		for {
			createOrUpdate := GetRandomNumInRange(-100, 1200)
			err := ls.executeUpdate(createOrUpdate < 0)
			if err != nil {
				ls.reportError(fmt.Errorf("failed to execute load simulation %w", err))
			}

			// now decide a random number of seconds to sleep and execute again
			randomSleepSeconds := GetRandomNumInRange(0, 4)
			time.Sleep(time.Duration(randomSleepSeconds) * time.Second)
		}
	}()
}

func (ls *LoadSimulator) executeUpdate(createMode bool) error {
	if createMode {
		err := ls.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DefaultBucketName))

			newPwd := uuid.NewString()[:10]
			encrypter, err := ls.encryptorFunc(newPwd)
			if err != nil {
				return fmt.Errorf("failed to create new encryptor %w", err)
			}

			key, err := crypt.CalculateHash([]byte(uuid.NewString()[:GetRandomNumInRange(4, 15)]))
			if err != nil {
				return err
			}

			v := []byte(gofakeit.Sentence(GetRandomNumInRange(50, 500)))
			encrypted, err := encrypter.Encrypt(v)
			if err != nil {
				return fmt.Errorf("failed to encrypt %w", err)
			}

			err = b.Put([]byte(key), encrypted)
			if err != nil {
				return fmt.Errorf("failed to put val %w", err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	}

	totalKeys, err := GetKeysCount(ls.db)
	if err != nil {
		return fmt.Errorf("faile to get total keys %w", err)
	}

	if totalKeys == 0 {
		return nil
	}

	randomIndex := GetRandomNumInRange(0, int(totalKeys)-1)

	err = ls.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DefaultBucketName))
		c := b.Cursor()

		index := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if index == randomIndex {
				for _, exception := range ls.exceptions {
					if bytes.Equal(k, []byte(exception)) {
						break
					}
				}

				newPwd := uuid.NewString()[:10]
				encrypter, err := ls.encryptorFunc(newPwd)
				if err != nil {
					return fmt.Errorf("failed to create new encryptor %w", err)
				}

				// either increase the length, or decrease the length
				words := GetRandomNumInRange(-100, 600)
				if words < 0 {
					// reduce by 300 bytes
					reduceFactor := 300
					if len(v) <= reduceFactor+1 {
						// delete
						err = b.Delete(k)
						if err != nil {
							return err
						}

						return nil
					}

					v = v[:len(v)-300]
				}

				v = append(v, []byte(gofakeit.Sentence(words))...)
				encrypted, err := encrypter.Encrypt(v)
				if err != nil {
					return fmt.Errorf("failed to encrypt %w", err)
				}

				err = b.Put(k, encrypted)
				if err != nil {
					return fmt.Errorf("failed to update val %w", err)
				}

				break
			}

			index++
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func GetRandomNumInRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func GenerateRandomNotes() []soul.Note {
	numOfNotes := GetRandomNumInRange(1, 80)

	var notes []soul.Note
	for i := 0; i < numOfNotes; i++ {
		notes = append(notes, soul.Note{
			Version: 1,
			ID:      uuid.NewString(),
			Text:    soul.NewBindingFromString(gofakeit.Sentence(GetRandomNumInRange(10, 15000))),
		})
	}

	return notes
}

func GetKeysCount(db *bolt.DB) (uint64, error) {
	var total uint64 = 0
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DefaultBucketName))

		var count uint64 = 0
		err := b.ForEach(func(k, v []byte) error {
			count++

			return nil
		})
		if err != nil {
			return err
		}

		total = count

		return nil
	})

	if err != nil {
		return 0, err
	}

	return total, nil
}

func NewLoadSimulator(db *bolt.DB, exceptions []string, encryptorFunc func(key string) (soul.Encrypter, error), reportErr func(error)) (*LoadSimulator, error) {
	// calculate hash of exceptions
	var exceptionsHash []string

	for _, exception := range exceptions {
		// first add as it is
		tmp, err := crypt.CalculateStringHash(exception)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash")

		}
		exceptionsHash = append(exceptionsHash, tmp)

		for i := 0; i < 1000; i++ {
			tmp, err := crypt.CalculateStringHash(fmt.Sprintf("%s%d", exception, i))
			if err != nil {
				return nil, fmt.Errorf("failed to calculate hash")
			}

			exceptionsHash = append(exceptionsHash, tmp)
		}
	}

	return &LoadSimulator{
		reportError:   reportErr,
		db:            db,
		exceptions:    exceptionsHash,
		encryptorFunc: encryptorFunc,
	}, nil
}
