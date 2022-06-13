package utils

import "encoding/binary"

// Uint32ToBytes converts uint32 to a byte slice
func Uint32ToBytes(l uint32) (res []byte) {
	res = make([]byte, 4)
	binary.LittleEndian.PutUint32(res, l)
	return
}
