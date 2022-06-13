package delta

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calculateDiff(t *testing.T, blockSize int, fileA, FileB []byte) ([]*BlockSignature, map[int]*Delta) {
	// generate signature for file A
	fileAbuf := bytes.NewReader(fileA)
	signatures, err := GenerateSignatures(fileAbuf, blockSize)
	require.NoError(t, err)

	// generate delta for file B using the signatures of fileA
	fileBBuf := bytes.NewBuffer(FileB)
	delta, err := GenerateDelta(fileBBuf, blockSize, signatures)
	require.NoError(t, err)

	return signatures, delta
}

func matchDiff(t *testing.T, expected map[int][]byte, delta map[int]*Delta) {
	for i, expect := range expected {
		actual, ok := delta[i]
		if !assert.Equalf(t, true, ok, "No matching delta for %d", i) {
			continue
		}

		// fmt.Printf("%s <=> %s\n", string(expect), string(actual.Data))
		assert.Equal(t, string(expect), string(actual.Data))
	}
}

func TestEqual(t *testing.T) {
	a := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")
	b := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")

	_, delta := calculateDiff(t, 1<<4, a, b)
	matchDiff(t, map[int][]byte{
		0: make([]byte, 0),
		1: make([]byte, 0),
		2: make([]byte, 0),
		3: make([]byte, 0),
	}, delta)
}

func TestChunkChange(t *testing.T) {
	a := []byte("When summertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("When summertime rolls in and the days hot enough that you need to cool off from the blazing heat")

	expect := map[int][]byte{
		3: []byte(" days hot en"),
	}

	_, delta := calculateDiff(t, 1<<4, a, b)
	matchDiff(t, expect, delta)
}
