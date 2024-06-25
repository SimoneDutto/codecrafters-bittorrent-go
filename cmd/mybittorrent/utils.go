package main

import (
	"crypto/sha1"
	"encoding/hex"
)

func peekUntil(s string, start int, charEnd rune) int {
	for i := start; i < len(s); i++ {
		if s[i] == byte(charEnd) {
			return i
		}
	}
	panic("cannot find peek char")
}

func calcSha1(bv []byte) string {
	hasher := sha1.New()
	hasher.Write(bv)
	return hex.EncodeToString(hasher.Sum(nil))
}
