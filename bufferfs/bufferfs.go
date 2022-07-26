// bufferfs is responsible for storing contents on an encrypted buffer
package bufferfs

import (
	"encoding/binary"
	"fmt"
)

// Buffer represents a slice of blocks
type Buffer []Block

// Block is a 4 KB buffer
type Block [BlockSizeBytes]byte

const BlockSizeBytes = 4096

// Store stores the value inside the buffer
func Store(buffer *Buffer, val []byte) error {
	if len(val) > len(*buffer) {
		return fmt.Errorf("buffer is smaller than value to be written")
	}

	var blocks []Block

	var tmp []byte
	binary.LittleEndian.PutUint64(tmp, uint64(len(val)))
	copy(blocks[0][:], tmp)

	blocks = append(blocks, CreateBlocks(val)...)
	copy(*buffer, blocks)

	return nil
}

// Get fetches the value from the buffer
func Get(buffer *Buffer) ([]byte, error) {

}

// CreateBlocks creates blocks from a byte array
func CreateBlocks(val []byte) []Block {
	var blocks []Block
	numOfBlocks := len(val) / BlockSizeBytes
	if len(val)%BlockSizeBytes != 0 {
		numOfBlocks++
	}

	for i := 0; i < numOfBlocks; i++ {
		block := Block{}

		upperIndex := (i + 1) * BlockSizeBytes
		if len(val[i*BlockSizeBytes:]) < BlockSizeBytes {
			upperIndex = len(val)
		}

		copy(block[:], val[i*BlockSizeBytes:upperIndex])
		blocks = append(blocks, block)
	}

	return blocks
}
