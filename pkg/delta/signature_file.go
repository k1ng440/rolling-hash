package delta

import (
	"encoding/gob"
	"errors"
	"os"
)

func WriteSignaturesToFile(filename string, sig chan *BlockSignature) error {
	signatures := make([]*BlockSignature, 0)

	for s := range sig {
		signatures = append(signatures, s)
	}

	if len(signatures) == 0 {
		return errors.New("can not write empty signatures to file")
	}

	fi, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fi.Close()

	g := gob.NewEncoder(fi)
	return g.Encode(signatures)
}

func ReadSignaturesFromFile(filename string) ([]*BlockSignature, error) {
	var res []*BlockSignature

	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	g := gob.NewDecoder(fi)
	if err := g.Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}
