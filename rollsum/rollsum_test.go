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

func TestRoll(t *testing.T) {
	// roll := New(100)
	// roll.Roll([]byte("Lorem ipsum dolor sit amet consectetur adipisicing elit"))
	// fmt.Printf("sum32=0x%x sum=0x%x", roll.Sum32(), roll.Sum([]byte{}))
}

// TestRollingHash tests that incrementally calculated signatures arrive to the same
// value as the full block signature.
// func TestRollingHash(t *testing.T) {
// 	roll := New(1024 * 4)
// 	roll.Write([]byte("abcd")) // file's content in server
// 	target := roll.Sum32()

// 	reader := bytes.NewReader([]byte("aabcdbbabcdddf")) // new file's content in client

// 	delta := make([]byte, 0)
// 	rolling := false
// 	offset := int64(0)

// 	for {
// 		buffer := make([]byte, 4) // block size of 4
// 		n, err := reader.ReadAt(buffer, offset)

// 		block := buffer[:n]
// 		if rolling {
// 			fmt.Println(string(block[n-1]))
// 			roll.Roll(block[n-1])
// 		} else {
// 			roll.Reset()
// 			roll.Write(block)
// 		}

// 		if roll.Sum32() == target {
// 			if err == io.EOF {
// 				break
// 			}

// 			rolling = false
// 			offset += int64(n - 1)
// 		} else {
// 			if err == io.EOF {
// 				delta = append(delta, block...)
// 				break
// 			}

// 			rolling = true
// 			delta = append(delta, roll.removed)
// 			offset++
// 		}

// 		assert.NoError(t, err)
// 	}

// 	assert.Equal(t, []byte("aabbddf"), delta)
// }

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
		roll.Roll(v)
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
			roll.Roll(db[len(db)-1])

			if !assert.Equal(t, classic([]byte(g.data)), roll.Sum32(), "invalid Roll") {
				fmt.Printf("orginal: %s\ncurrent window: %v\nblockSize: %d\n\n", string(db), string(roll.Window()), len(roll.window))
			}
		})
	}
}
