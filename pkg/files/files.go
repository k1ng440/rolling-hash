package files

import (
	"errors"
	"io"
	"os"
	"encoding/gob"

	"github.com/k1ng440/rolling-hash/pkg/delta"
)

func ReadFile(filename string) (io.Reader, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}

	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return fi, nil
}


func WriteDelta(filename string, data map[int]*delta.Delta) error {
	fi, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fi.Close()

	g := gob.NewEncoder(fi)
	return g.Encode(data)
}

