package delta

import (
	"errors"
	"fmt"

	"github.com/k1ng440/rolling-hash/internal/utils"
	"github.com/k1ng440/rolling-hash/rollsum"
)

const (
	DefaultBlockSize = 1024 * 6 // 6kb block size
)

type BlockSignature struct {
	// Block index
	Index int
	// Strong checksum
	Strong []byte
	// rsync rolling checksum
	Weak uint32
	// Error is used to report the error reading the file or calculating checksums
	Error error

	// BlockData is used for debugging purpose only
	BlockData []byte
}

// GenerateSignatures calculates the signature of given target
func GenerateSignatures(target chan *utils.File, blockSize int) ([]*BlockSignature, error) {
	result := make([]*BlockSignature, 0)
	strongHasher := utils.NewHasher()
	weakHasher := rollsum.New(blockSize)

	if blockSize == 0 {
		return nil, errors.New("blockSize must be greater than 0")
	}

	for buf := range target {
		if buf.Error != nil {
			// forward the error
			return nil, buf.Error
		}

		fmt.Printf("sig %d => %d = '%s'\n", buf.BlockIndex, len(buf.Block), string(buf.Block))

		stronghash := strongHasher.MakeHash(buf.Block)
		weakHasher.Reset()
		weakHasher.Write(buf.Block)

		result = append(result, &BlockSignature{
			Strong:    stronghash,
			Weak:      weakHasher.Sum32(),
			Index:     buf.BlockIndex,
			BlockData: buf.Block,
		})
	}

	return result, nil
}
