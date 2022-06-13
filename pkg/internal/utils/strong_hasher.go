package utils

import (
	"crypto/md5"
	"hash"
)

func NewHasher() *MD5Hasher {
	return &MD5Hasher{
		hasher: md5.New(),
	}
}

type MD5Hasher struct {
	hasher hash.Hash
}

func (s *MD5Hasher) MakeHash(b []byte) []byte {
	s.hasher.Reset()
	s.hasher.Write(b)
	return s.hasher.Sum(nil)
}
