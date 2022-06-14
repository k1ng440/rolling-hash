package rollsum

import (
	"fmt"
	"hash/adler32"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var golden = []struct {
	sum  uint32
	data string
}{
	{0x00620062, "a"},
	{0x012600c4, "ab"},
	{0x024d0127, "abc"},
	{0x03d8018b, "abcd"},
	{0x05c801f0, "abcde"},
	{0x081e0256, "abcdef"},
	{0x0adb02bd, "abcdefg"},
	{0x0e000325, "abcdefgh"},
	{0x118e038e, "abcdefghi"},
	{0x158603f8, "abcdefghij"},
	{0x2b6f31f4, "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Duis eu purus vitae turpis placerat ullamcorper. Quisque ac libero eget nulla"},
}

func classic(data []byte) uint32 {
	a := adler32.New()
	a.Write(data)
	return a.Sum32()
}

func TestBasic(t *testing.T) {
	roll := New(1024 * 1024)
	roll.Write([]byte("abcdefghi"))
	require.Equal(t, 9, roll.Size())
	require.Equal(t, uint32(0x118e038e), roll.Sum32())

	old := roll.Sum32()
	for _, v := range []byte("abcdefghi") {
		roll.Rotate(v)
	}
	require.Equal(t, old, roll.Sum32())
}

func TestGolden(t *testing.T) {
	for _, g := range golden {
		t.Run(fmt.Sprintf("0x%x", g.sum), func(t *testing.T) {
			db := []byte(g.data)

			c := classic(db)
			if !assert.Equal(t, c, g.sum, "test golden") {
				return
			}

			roll := New(1024 * 1024)
			roll.Write(append([]byte("\x00"), db[:len(db)-1]...))
			roll.Rotate(db[len(db)-1])
			assert.Equal(t, classic([]byte(g.data)), roll.Sum32())
		})
	}
}

func TestRollInAndOut(t *testing.T) {
	a := []byte("Adler-32 is a checksum")
	roll := New(1 << 16)
	for _, x := range a {
		roll.In(x)
	}
	require.Equal(t, classic(a), roll.Sum32())

	roll.Out()
	a = a[1:]
	assert.Equal(t, classic(a), roll.Sum32())

	roll.Out()
	a = a[1:]
	assert.Equal(t, classic(a), roll.Sum32())

	roll.Out()
	a = a[1:]
	assert.Equal(t, classic(a), roll.Sum32())

	roll.Out()
	a = a[1:]
	assert.Equal(t, classic(a), roll.Sum32())

	roll.Out()
	a = a[1:]
	assert.Equal(t, classic(a), roll.Sum32())
}

func TestSum(t *testing.T) {
	testData := []byte("abcdefghi")
	roll := New(1024 * 1024)
	roll.Write(testData)
	sum := roll.Sum([]byte{})

	adler := adler32.New()
	adler.Write(testData)

	assert.Equal(t, adler.Sum([]byte{}), sum)
}
