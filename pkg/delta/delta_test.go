package delta

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/k1ng440/rolling-hash/internal/utils"
	"github.com/stretchr/testify/require"
)

func dummyUtilFile(blockSize int, data []byte) chan *utils.File {
	c := make(chan *utils.File)
	tmp := make([]byte, len(data))
	copy(tmp, data)

	go func() {
		defer close(c)

		idx := 0
		for len(tmp) > 0 {
			end := blockSize

			if len(tmp) < end {
				end = len(tmp)
			}

			c <- &utils.File{
				BlockIndex: idx,
				Block:      tmp[:end],
				BlockSize:  len(tmp[:end]),
			}

			idx++

			tmp = tmp[end:]
		}
	}()

	return c
}

func calculateDiff(t *testing.T, blocksize int, fileA, FileB []byte) ([]*BlockSignature, map[int]*Delta) {
	// generate signature for file A
	fileAChan := dummyUtilFile(blocksize, fileA)
	signatures, err := GenerateSignatures(fileAChan, blocksize)
	require.NoError(t, err)

	// generate delta for file B using the signatures of fileA
	fileBBuf := bytes.NewBuffer(FileB)
	delta, err := GenerateDelta(fileBBuf, blocksize, signatures)
	require.NoError(t, err)

	return signatures, delta
}

func TestEqual(t *testing.T) {
	a := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")
	b := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")

	_, delta := calculateDiff(t, 30, a, b)
	for _, d := range delta {
		fmt.Printf("%#v\n", d)
	}
}
