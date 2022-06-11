// package rollinghash is a basic implementation of rolling hash based file diffing algorithm described in rsync phd thesis
// https://www.samba.org/~tridge/phd_thesis.pdf
// https://www.andrew.cmu.edu/course/15-749/readings/required/cas/tridgell96.pdf
package rollinghash

import (
	"fmt"
	"io"
)

const (
	DefaultBlockSize = 1024 * 6 // 6kb block size
	mod              = DefaultBlockSize * 10
)

type BlockSignature struct {
	// Block index
	Index uint64
	// Strong checksum
	Strong []byte
	// rsync rolling checksum
	Weak uint32
	// Error is used to report the error reading the file or calculating checksums
	Error error
}

// GenerateSignatures calculates the signature of given target
func GenerateSignatures(target io.Reader, blockSize int, bs chan *BlockSignature) {
	defer close(bs)
	strongHasher := NewHasher()
	buf := make([]byte, blockSize)
	idx := uint64(0)

	if blockSize == 0 {
		blockSize = DefaultBlockSize
	}

	l := true
	for l {
		// ReadAtLeast read exactly *blockSize* bytes from target into buf
		// It returns the number of bytes copied and an error if fewer bytes were read.
		// The error is EOF only if no bytes were read
		n, err := io.ReadAtLeast(target, buf, blockSize)
		if err != nil {
			// reached the end of the file
			if err == io.EOF {
				return
			}

			// Catch all errors except ErrUnexpectedEOF
			if err != io.ErrUnexpectedEOF {
				bs <- &BlockSignature{
					Error: err,
				}
				return
			}

			// Unexpected EOF. last loop with last bit of remaining buffer
			l = false
		}

		block := buf[:n]
		strong := strongHasher.MakeHash(block)
		weak, _, _ := WeakChecksum(block)
		bs <- &BlockSignature{
			Strong: strong,
			Weak:   weak,
			Index:  idx,
		}

		fmt.Printf("signature: len=%d\np=%s\nweak=0x%x strong=0x%x\n------------------------------------------------\n\n", n, string(block), weak, string(strong))

		idx++
	}
}

// ReadSignature reads signatures from buffer
func ReadSignature(r io.Reader) {

}
