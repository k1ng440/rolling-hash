package delta

import (
	"bytes"
	"encoding/hex"
	"testing"
	"os"

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

		assert.Equalf(t, string(expect), string(actual.Literal), "Signature Index %d", i)
	}
}

func printDelta(t *testing.T, deltas map[int]*Delta) {
	if os.Getenv("PRINT") == "" {
		return
	}


	ds := make([]*Delta, len(deltas))
	for i, delta := range deltas {
		ds[i] = delta
	}

	for i, delta := range ds {
		t.Logf("Delta: Sig index: %d => missing=%v start=%d end=%d literal='%s'", i, delta.Missing, delta.Start, delta.End, string(delta.Literal))
	} 
}

func printSignatures(t *testing.T, signatures []*BlockSignature) {
	if os.Getenv("PRINT") == "" {
		return
	}

	for _, sig := range signatures {
		strong := hex.EncodeToString(sig.Strong[:])
		t.Logf("Signature: Sig index %d => weak=0x%d strong=%s data='%s'", sig.Index, sig.Weak, strong, string(sig.BlockData))
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

	_, delta := calculateDiff(t, 16, a, b)

	assertMissingDelta(t, 0, false, delta)
	assertDiff(t, map[int][]byte{
		0: make([]byte, 0),
	}, delta)
}

func TestEqual(t *testing.T) {
	a := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")
	b := []byte("Be yourself; everyone else is already taken. - Oscar Wilde")

	_, delta := calculateDiff(t, 16, a, b)

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
	a := []byte("When wintertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("When summertime rolls in and the days hot enough that you need to cool off from the blazing heat")

	_, delta := calculateDiff(t, 16, a, b)
	printDelta(t, delta)

	assertMissingDelta(t, 0, true, delta)
	assertMissingDelta(t, 2, true, delta)
	expect := map[int][]byte{
		1: []byte("When summertime "),
		3: []byte(" days hot en"),
	}
	assertDiff(t, expect, delta)
}

func TestChunkAddition(t *testing.T) {
	a := []byte("When summertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("When summertime rolls in and the days get hot en ..... new additionough that you need to cool off from the blazing heat")
	_, delta := calculateDiff(t, 16, a, b)
	printDelta(t, delta)

	expect := map[int][]byte{
		3: []byte(" ..... new addition"),
	}

	assertDiff(t, expect, delta)
}

func TestChunkRemoved(t *testing.T) {
	a := []byte("When summertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("rolls in and the days get hot enough that you ne rom the blazing heat")
	sig, delta := calculateDiff(t, 16, a, b)
	printSignatures(t, sig)
	printDelta(t, delta)

	assertMissingDelta(t, 0, true, delta)
	assertMissingDelta(t, 4, true, delta)
}

func TestChunkShift(t *testing.T) {
	
	a := []byte("When summertime rolls in and the days get hot enough that you need to cool off from the blazing heat")
	b := []byte("When summertim   e rolls in and the days get hot enough        that you need to cool off from the blazing heat")

	sig, delta := calculateDiff(t, 16, a, b)
	printSignatures(t, sig)
	printDelta(t, delta)

	expect := map[int][]byte{
		1: []byte("When summertim   e "),
		4: []byte("ough        that you ne"),
	}

	assertMissingDelta(t, 0, true, delta)
	assertMissingDelta(t, 3, true, delta)
	assertDiff(t, expect, delta)

}
