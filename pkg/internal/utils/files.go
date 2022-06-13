package utils

import (
	"io"
	"os"
)

type File struct {
	Block      []byte
	BlockSize  int
	BlockIndex int
	Error      error
}

// ReadFile reads the files and sends the block using channel
func ReadFile(filename string, blockSize int) chan *File {
	c := make(chan *File)

	reader, err := os.Open(filename)
	if err != nil {
		c <- &File{Error: err}
	}

	go func() {
		defer close(c)

		buf := make([]byte, blockSize)
		index := 0

		for {
			// ReadAtLeast read exactly *blockSize* bytes from target into buf
			// It returns the number of bytes copied and an error if fewer bytes were read.
			// The error is EOF only if no bytes were read
			n, err := io.ReadAtLeast(reader, buf, blockSize)
			if err != nil {
				if err == io.EOF {
					return
				}

				// catch and throw all errors except the unexpected EOF
				if err != io.ErrUnexpectedEOF {
					c <- &File{Error: err}
					return
				}
			}

			c <- &File{
				Block:      buf[:n],
				BlockSize:  n,
				BlockIndex: index,
			}

			index++

			if err == io.ErrUnexpectedEOF {
				// reached end of the file buffer
				return
			}
		}
	}()

	return c
}
