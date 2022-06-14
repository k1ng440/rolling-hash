package delta

import (
	"errors"
	"io"

	"github.com/k1ng440/rolling-hash/pkg/internal/utils"
	"github.com/k1ng440/rolling-hash/pkg/rollsum"
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
	// BlockData is used for debugging purpose
	BlockData []byte
}

// GenerateSignatures calculate signatures of given target by dividing them blocks
func GenerateSignatures(target io.Reader, blockSize int) ([]*BlockSignature, error) {
	result := make([]*BlockSignature, 0)
	strongHasher := utils.NewHasher()
	weakHasher := rollsum.New(blockSize)

	if blockSize == 0 {
		return nil, errors.New("blockSize must be greater than 0")
	}

	loop := true
	index := 0
	for loop {
		block := make([]byte, blockSize)
		n, err := io.ReadAtLeast(target, block, blockSize)
		if err != nil {
			// end of the file
			if err == io.EOF {
				break
			}

			// catch all error except io.ErrUnexpectedEOF
			if err != io.ErrUnexpectedEOF {
				return nil, err
			}

			// last loop because of the io.ErrUnexpectedEOF.
			loop = false
		}

		buf := block[:n]
		strongHash := strongHasher.MakeHash(buf)
		weakHasher.Reset()
		weakHasher.Write(buf)

		result = append(result, &BlockSignature{
			Strong: strongHash,
			Weak:   weakHasher.Sum32(),
			Index:  index,
			BlockData: buf,
		})

		index++
	}

	return result, nil
}
