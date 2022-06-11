package rollinghash_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	rollinghash "github.com/k1ng440/rolling-hash"
	"github.com/k1ng440/rolling-hash/internal/utils"
	"github.com/stretchr/testify/require"
)

type goldenFile struct {
	basisFile io.Reader
	newFile   io.Reader
	desc      string
}

func readFile(t *testing.T, fileName string) io.Reader {
	inputData, err := ioutil.ReadFile("testdata/" + fileName)
	require.NoError(t, err)
	return bytes.NewReader(inputData)
}

func newGolden(t *testing.T, fileBase string, desc string) *goldenFile {
	return &goldenFile{
		basisFile: readFile(t, fileBase+".old"),
		newFile:   readFile(t, fileBase+".new"),
		desc:      desc,
	}
}

func TestRollingHash(t *testing.T) {
	goldens := []*goldenFile{
		newGolden(t, "lorem-ipsum", "Lorem Ipsum"),
	}

	r := bytes.NewBuffer([]byte{})

	for _, golden := range goldens {
		t.Run(golden.desc, func(t *testing.T) {
			sig := make(chan *rollinghash.BlockSignature)
			go rollinghash.GenerateSignatures(golden.newFile, 0, sig)
			for c := range sig {
				require.NoError(t, c.Error)
				r.Write(utils.Uint32ToBytes(c.Weak))
				r.Write(c.Strong)
			}

		})
	}

}
