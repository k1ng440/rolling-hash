package files

import (
	"encoding/gob"
	"errors"
	"os"
	"github.com/k1ng440/rolling-hash/pkg/delta"
)

// WriteSignaturesToFile encodes byte slice using gob and writes to a file
// returns error if no signatures given or failed to open file
func WriteSignaturesToFile(filename string, signatures []*delta.BlockSignature) error {
	if len(signatures) == 0 {
		return errors.New("can not write empty signatures to file")
	}

	fi, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fi.Close()

	g := gob.NewEncoder(fi)
	return g.Encode(signatures)
}

// ReadSignaturesFromFile reads gob encoded signatures from file 
// returns error if file not found or contains invalid signatures
func ReadSignaturesFromFile(filename string) ([]*delta.BlockSignature, error) {
	var res []*delta.BlockSignature

	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	g := gob.NewDecoder(fi)
	if err := g.Decode(&res); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.New("the file does not contains valid signatures")
	}

	return res, nil
}
