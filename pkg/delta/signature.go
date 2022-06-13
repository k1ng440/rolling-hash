package delta

import (
	"errors"
	"io"

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

	// BlockData is used for debugging purpose
	BlockData []byte
}

// GenerateSignatures calculates the signature of given target
func GenerateSignatures(target io.Reader, blockSize int) ([]*BlockSignature, error) {
	result := make([]*BlockSignature, 0)
	strongHasher := utils.NewHasher()
	weakHasher := rollsum.New(blockSize)

	if blockSize == 0 {
		return nil, errors.New("blockSize must be greater than 0")
	}

	loop := true
	index := 0
	block := make([]byte, blockSize)
	for loop {

		n, err := io.ReadAtLeast(target, block, blockSize)
		if err != nil {
			if err == io.EOF {
				return nil, err
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
			// BlockData: buf,
		})

		index++
	}

	return result, nil
}
