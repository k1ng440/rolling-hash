package files

import (
	"path/filepath"
	"testing"
	"os"

	"github.com/k1ng440/rolling-hash/pkg/delta"
	"github.com/stretchr/testify/assert"
)


func TestWriteSignatureToFile(t *testing.T) {
	sigPath := filepath.Join(t.TempDir(), "signature.sig")


	signatures := make([]*delta.BlockSignature, 0)
	for i := 0; i < 10; i++ {
		signatures = append(signatures, &delta.BlockSignature{
			Index: i,
		})
	}

	err := WriteSignaturesToFile(sigPath, signatures)
	assert.NoError(t, err)
	assert.FileExists(t, sigPath)
}

func TestWriteBadSignatureTofile(t *testing.T) {
	sigPath := filepath.Join(t.TempDir(), "signature-err.sig")

	err := WriteSignaturesToFile(sigPath, []*delta.BlockSignature{})
	assert.Error(t, err)
	assert.NoFileExists(t, sigPath)
}

func TestReadSignature(t *testing.T) {
	t.Run("read signature from file", func(t *testing.T) {
		sigPath := filepath.Join(t.TempDir(), "signature-read.sig")
		t.Run("write", func(t *testing.T) { 
			signatures := make([]*delta.BlockSignature, 0)
			for i := 0; i < 10; i++ {
				signatures = append(signatures, &delta.BlockSignature{
					Index: i,
				})
			}

			err := WriteSignaturesToFile(sigPath, signatures)
			assert.NoError(t, err)
			assert.FileExists(t, sigPath)
		})

		t.Run("read", func(t *testing.T) {
			sigs, err := ReadSignaturesFromFile(sigPath)
			assert.NoError(t, err)
			assert.Len(t, sigs, 10)
		})
	})

	t.Run("bad signature", func(t *testing.T) {
		sigPath := filepath.Join(t.TempDir(), "signature-read.sig")
		f, err := os.OpenFile(sigPath, os.O_RDWR|os.O_CREATE, 0755)
		if assert.NoError(t, err) {
			f.Close()
		}

		_, err = ReadSignaturesFromFile(sigPath)
		assert.Error(t, err)
	})
}
