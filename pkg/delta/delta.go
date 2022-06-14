package delta

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/k1ng440/rolling-hash/pkg/internal/utils"
	"github.com/k1ng440/rolling-hash/pkg/rollsum"
)

type Delta struct {
	Start          int
	End            int
	Missing        bool
	Data           []byte
	SignatureIndex int
}

type signatureMap map[uint32][]*BlockSignature

func (sm signatureMap) initialize(sigs []*BlockSignature) {
	for _, sig := range sigs {
		if sm[sig.Weak] == nil {
			sm[sig.Weak] = make([]*BlockSignature, 0)
		}

		sm[sig.Weak] = append(sm[sig.Weak], sig)
	}
}

// match compares the weak hashes and then confirm with strong hashes
// returns signature index if found otherwise -1
func (sm signatureMap) match(weakHash uint32, window []byte) int {
	if sigs, ok := sm[weakHash]; ok {
		strongHasher := utils.NewHasher()
		for _, sig := range sigs {
			// Confirm the signature between 2 block are equal using strong hash
			// in our case we are using md5
			if bytes.Equal(sig.Strong, strongHasher.MakeHash(window)) {
				// strong hash matched
				return sig.Index
			}
		}
	}

	// no matching signature found
	return -1
}

// verifyIntegrity checks for removed block and add them the missing deltas map
func verifyIntegrity(blockSize int, sigs []*BlockSignature, deltas map[int]*Delta) {
	for _, sig := range sigs {
		if _, ok := deltas[sig.Index]; !ok {
			// Add the missing blocks
			deltas[sig.Index] = &Delta{
				Missing:        true,
				Start:          sig.Index * blockSize,
				End:            (sig.Index * blockSize) + blockSize,
				SignatureIndex: sig.Index,
			}
		}
	}
}

// GenerateDelta generates diff by calculating and matching signature of given buffer
func GenerateDelta(reader io.Reader, blockSize int, signatures []*BlockSignature) (map[int]*Delta, error) {
	if blockSize == 0 {
		return nil, errors.New("blockSize must be greater than 0")
	}

	if len(signatures) == 0 {
		return nil, errors.New("can not calculate delta from empty signature")
	}

	// Initialize the signature lookup map
	sigMap := make(signatureMap)
	sigMap.initialize(signatures)

	roll := rollsum.New(DefaultBlockSize)
	buf := bufio.NewReader(reader)
	result := make(map[int]*Delta)
	diff := make([]byte, 0, blockSize)
	eof := false // End of file
	for !eof {
		// read single byte from the buffer
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				if roll.Size() == 0 {
					// Reached the end of the file and rolling hash window is empty
					break
				}

				// mark as last loop
				eof = true
			} else {
				return nil, err
			}
		}

		// Add byte to rolling hash if we have not reached end of the file
		// This condition is to prevent adding nil to rolling hash window
		if !eof {
			roll.In(b)

			// Build up the rolling hash window to match the block size
			// Exception: rolling hash window can be smaller if reached the EOF
			if roll.Size() < blockSize {
				continue
			}
		}

		// Match signature of rolling hash
		index := sigMap.match(roll.Sum32(), roll.Window())
		if index == -1 { // no match
			// Remove the oldest byte from the rolling hash window and store it in diff
			roll.Out()
			diff = append(diff, roll.Removed())
			continue
		}

		// Add the matching delta
		result[index] = &Delta{
			Start:          index * blockSize,
			End:            (index * blockSize) + blockSize,
			Data:           diff,
			SignatureIndex: index,
		}

		// Reset rollsum for next window
		roll.Reset()
		diff = make([]byte, 0, blockSize)
	}

	// Verify the integrity of buffer and missing blocks to result
	verifyIntegrity(blockSize, signatures, result)

	return result, nil
}
