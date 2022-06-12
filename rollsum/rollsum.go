// rollsum is modd on rsync Adler32 rolling checksum
// https://rsync.samba.org/tech_report/node3.html
// RFC https://datatracker.ietf.org/doc/html/rfc1950
package rollsum

import (
	"errors"
	"hash"
	"hash/adler32"
)

const (
	// modulo 2^16 ( largest prime smaller than 65536 )
	// defined in RFC 1950
	mod = 65521
)

type Rollsum struct {
	a, b int32

	window    []byte
	blockLen  int32 // block size
	windowCap int
	hasher    hash.Hash32
	removed   byte
}

func New(windowCap int) *Rollsum {
	return &Rollsum{
		a:         1,
		window:    make([]byte, 0),
		windowCap: windowCap,
		hasher:    adler32.New(),
	}
}

func (r *Rollsum) Reset() {
	r.a = 1
	r.b = 0
	r.blockLen = 0
	r.window = make([]byte, 0)
	r.hasher.Reset()
}

func (r *Rollsum) Write(block []byte) (int, error) {
	if len(block) == 0 {
		return 0, nil
	}

	if len(r.window) > 0 {
		return 0, errors.New("adding to initialized window is not allowed")
	}

	// TODO sliding window
	r.window = append(r.window, block...)

	r.hasher.Reset()
	r.hasher.Write(r.window)
	sum := r.hasher.Sum32()

	r.a = int32(sum & 0xffff)
	r.b = int32(sum>>16) & 0xffff
	r.blockLen = int32(len(block))

	return int(r.blockLen), nil
}

func (r *Rollsum) slide(b byte) (leave, enter int32) {
	r.removed = r.window[0]
	r.window = append(r.window[1:], b)
	enter = int32(b)
	leave = int32(r.removed)
	return
}

// Roll adds a byte to checksum
func (r *Rollsum) Roll(b byte) {
	if len(r.window) == 0 {
		return
	}

	leave, enter := r.slide(b)
	r.a = (((r.a + enter - leave) % mod) + mod) % mod
	r.b = (((r.b - r.blockLen*leave - 1 + r.a) % mod) + mod) % mod
}

func (r *Rollsum) Window() []byte {
	return r.window
}

func (r *Rollsum) WindowCap() int {
	return r.windowCap
}

func (r *Rollsum) Size() int {
	return int(r.blockLen)
}

func (r *Rollsum) Sum32() uint32 {
	return uint32(r.b<<16 | r.a)
}

func (r *Rollsum) Sum(in []byte) []byte {
	s := r.Sum32()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}
