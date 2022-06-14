package files

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/k1ng440/rolling-hash/pkg/delta"
	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	tmp := path.Join(t.TempDir(), "tmp")
	data := []byte("The quick brown fox jumps over the lazy dog")
	ioutil.WriteFile(tmp, data, 0644)

	reader, err := ReadFile(tmp)
	assert.NoError(t, err)

	res, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, res)
}

func TestWriteDelta(t *testing.T) {
	deltaPath := path.Join(t.TempDir(), "signature.delta")

	deltas := make(map[int]*delta.Delta)
	for i := 0; i < 10; i++ {
		deltas[i] = &delta.Delta{}
	}

	err := WriteDelta(deltaPath, deltas)
	assert.NoError(t, err)
}
