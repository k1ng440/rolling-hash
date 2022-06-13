// Package rollsum implements rsync Adler32 rolling checksum
// https://rsync.samba.org/tech_report/node3.html
// https://www.samba.org/~tridge/phd_thesis.pdf
// https://www.andrew.cmu.edu/course/15-749/readings/required/cas/tridgell96.pdf
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
	a, b uint32

	window    []byte
	windowLen uint32 // window size
	windowCap int
	hasher    hash.Hash32
	removed   byte
}

// New returns a new instance of Rollsum
// Note: Rollsum is unsafe and should not be used in go routines without mutual exclusion (sync.Mutex)
func New(windowCap int) *Rollsum {
	return &Rollsum{
		a:         1,
		window:    make([]byte, 0, windowCap),
		windowCap: windowCap,
		hasher:    adler32.New(),
	}
}

func (r *Rollsum) Reset() {
	r.a = 1
	r.b = 0
	r.windowLen = 0
	r.window = make([]byte, 0, r.windowCap)
	r.hasher.Reset()
}

// Write writes the initial window.
// the rolling does not occur in this method
// once the window has been initialized use Roll() method
func (r *Rollsum) Write(block []byte) (int, error) {
	if len(block) == 0 {
		return 0, nil
	}

	if len(append(r.window, block...)) > r.windowCap {
		return 0, errors.New("window cap has reached")
	}

	defer r.updateBlockLen()

	r.window = append(r.window, block...)
	r.hasher.Reset()
	r.hasher.Write(r.window)
	sum := r.hasher.Sum32()

	r.a = sum & 0xffff
	r.b = (sum >> 16) & 0xffff
	return int(r.windowLen), nil
}

// Roll adds a byte to checksum
func (r *Rollsum) Rotate(b byte) {
	if len(r.window) == 0 {
		return
	}

	leave, enter := r.circle(b)
	r.a = (((r.a + enter - leave) % mod) + mod) % mod
	r.b = (((r.b - r.windowLen*leave - 1 + r.a) % mod) + mod) % mod
}

// In adds the given byte to checksum
func (r *Rollsum) In(in byte) {
	defer r.updateBlockLen()
	r.window = append(r.window, in)

	// a = (a + in) % mod
	r.a = (r.a + uint32(in)) % mod
	// b = (b + a) % mod
	r.b = (r.b + r.a) % mod
}

// Out removes the oldest byte from checksum
func (r *Rollsum) Out() {
	if len(r.window) == 0 {
		r.Reset()
		return
	}

	r.removed = r.window[0]
	r.window = r.window[1:]

	r.a = (r.a - uint32(r.removed) + mod) % mod
	r.b = (r.b - ((r.windowLen * uint32(r.removed)) % mod) - 1 + mod) % mod
	r.updateBlockLen()
}

// Window returns current window used to generate checksum
func (r *Rollsum) Window() []byte {
	return r.window
}

// Removed returns the last removed byte from the window
func (r *Rollsum) Removed() byte {
	return r.removed
}

// Size returns underneath block size
func (r *Rollsum) Size() int {
	return int(r.windowLen)
}

// Sum32 returns an uint32 checksum of the working window
func (r *Rollsum) Sum32() uint32 {
	return uint32(r.b<<16 | r.a)
}

// Sum appends checksum of current window to given byte slice
func (r *Rollsum) Sum(in []byte) []byte {
	s := r.Sum32()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}

func (r *Rollsum) updateBlockLen() {
	r.windowLen = uint32(len(r.window)) % mod
}

func (r *Rollsum) circle(b byte) (leave, enter uint32) {
	r.removed = r.window[0]
	r.window = append(r.window[1:], b)
	enter = uint32(b)
	leave = uint32(r.removed)
	return
}
