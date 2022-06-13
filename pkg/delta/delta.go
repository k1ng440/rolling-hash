package delta

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/k1ng440/rolling-hash/internal/utils"
	"github.com/k1ng440/rolling-hash/rollsum"
)

type Delta struct {
	Start          int
	Offset         int
	Missing        bool
	Data           []byte
	SignatureIndex int
}

type signatureMap map[uint32][]*BlockSignature

// convert signature slice to map because hashes may correlate many unique hashes
func buildHashTable(sigs []*BlockSignature) signatureMap {
	res := make(signatureMap)

	for _, sig := range sigs {
		if res[sig.Weak] == nil {
			res[sig.Weak] = make([]*BlockSignature, 0)
		}

		res[sig.Weak] = append(res[sig.Weak], sig)
	}

	return res
}

// matchSignature compares weak and strong
// returns positive number if a matching signature found
// returns -1 on no match
func matchSignature(signatures signatureMap, weakHash uint32, data []byte) int {
	if sigs, ok := signatures[weakHash]; ok {
		strongHasher := utils.NewHasher()

		for _, sig := range sigs {
			// confirm the signature between 2 block are equal using strong hash
			// in our case we are using md5
			if bytes.Equal(sig.Strong, strongHasher.MakeHash(data)) {
				// strong hash matched
				return sig.Index
			}
		}
	}

	// no signature match
	return -1
}

func integrityCheck(blockSize int, sigs []*BlockSignature, deltas map[int]*Delta) {
	for _, sig := range sigs {
		if _, ok := deltas[sig.Index]; !ok {
			// missing or changed
			deltas[sig.Index] = &Delta{
				Missing:        true,
				Start:          sig.Index * blockSize,
				Offset:         (sig.Index * blockSize) + blockSize,
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

	result := make(map[int]*Delta)

	// build a map to reduce compute complexity during the weak signature matching
	sigLookup := buildHashTable(signatures)

	roll := rollsum.New(DefaultBlockSize)
	buf := bufio.NewReader(reader)
	tmpData := make([]byte, 0)
	for {
		// read single byte from the buffer
		c, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		// build the initial block
		if roll.Size() < blockSize {
			// expends the window until desired
			roll.In(c)
		} else {
			// add new byte and remove the oldest from window
			roll.Rotate(c)
			tmpData = append(tmpData, roll.Removed())
		}

		// match signature of the window
		index := matchSignature(sigLookup, roll.Sum32(), roll.Window())
		if index == -1 {
			continue
		}

		// add the matching block
		result[index] = &Delta{
			Offset:         (index * blockSize) + blockSize,
			Start:          index * blockSize,
			Data:           tmpData,
			SignatureIndex: index,
		}

		// reset rollsum for next match
		roll.Reset()
		tmpData = make([]byte, 0)
	}

	// add missing blocks to result
	integrityCheck(blockSize, signatures, result)

	return result, nil
}
