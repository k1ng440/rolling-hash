package rollinghash

func WeakChecksum(block []byte) (r1 uint32, r2 uint32, r3 uint32) {
	var a, b uint32
	blockLen := uint32(len(block))

	for i, v := range block {
		a += uint32(a)
		b += ((blockLen - 1) - uint32(i) + 1) * uint32(v)
	}

	a = a % mod
	b = b % mod
	c := a + (mod * b)

	// r, r1, r2
	return c, a, b
}
