package delta

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calculateDiff(t *testing.T, blockSize int, fileA, FileB []byte) ([]*BlockSignature, map[int]*Delta) {
	// Generate signature for file A
	fileABuf := bytes.NewReader(fileA)
	signatures, err := GenerateSignatures(fileABuf, blockSize)
	require.NoError(t, err)

	// Generate delta for file B using the signatures of fileA
	fileBBuf := bytes.NewBuffer(FileB)
	delta, err := GenerateDelta(fileBBuf, blockSize, signatures)
	require.NoError(t, err)

	return signatures, delta
}

func assertDiff(t *testing.T, expected map[int][]byte, delta map[int]*Delta) {
	for i, expect := range expected {
		actual, ok := delta[i]
		if !assert.Equalf(t, true, ok, "No matching delta for %d", i) {
			continue
		}

		assert.Equal(t, string(expect), string(actual.Data))
	}
}

func assertMissingDelta(t *testing.T, index int, expected bool, delta map[int]*Delta) bool {
	d, ok := delta[index]
	if !assert.Equalf(t, true, ok, "No matching delta for %d", index) {
		return false
	}

	return assert.Equal(t, expected, d.Missing)
}

// Test for small block
func TestEndOfDeltaFile(t *testing.T) {
	a := []byte("Be yourself")
	b := []byte("Be yourself")

	_, delta := calculateDiff(t, 1<<4, a, b)

	assertMissingDelta(t, 0, false, delta)

	assertDiff(t, map[int][]byte{
		0: make([]byte, 0),
	}, delta)
}

func TestEqual(t *testing.T) {
	a := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")
	b := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")

	_, delta := calculateDiff(t, 1<<4, a, b)

	assertMissingDelta(t, 0, false, delta)
	assertMissingDelta(t, 1, false, delta)
	assertMissingDelta(t, 2, false, delta)
	assertMissingDelta(t, 3, false, delta)

	assertDiff(t, map[int][]byte{
		0: make([]byte, 0),
		1: make([]byte, 0),
		2: make([]byte, 0),
		3: make([]byte, 0),
	}, delta)
}

func TestChunkChange(t *testing.T) {
	a := []byte("When summertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("When summertime rolls in and the days hot enough that you need to cool off from the blazing heat")

	_, delta := calculateDiff(t, 1<<4, a, b)

	expect := map[int][]byte{
		3: []byte(" days hot en"),
	}
	assertDiff(t, expect, delta)

}
