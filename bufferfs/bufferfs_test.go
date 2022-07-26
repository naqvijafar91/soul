package bufferfs_test

import (
	"soul/bufferfs"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func TestCreateBlocks(t *testing.T) {
	t.Parallel()

	val := []byte(gofakeit.Paragraph(7, 50, 5, " "))

	blocks := bufferfs.CreateBlocks(val)
	var fetched []byte
	for _, block := range blocks {
		for _, valByte := range block {
			fetched = append(fetched, valByte)
		}
	}

	for i := 0; i < len(val); i++ {
		assert.Equal(t, val[i], fetched[i])
	}
}
