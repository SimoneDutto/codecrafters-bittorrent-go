package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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

func extractInfo(v interface{}) (string, map[string]interface{}) {
	metaInfo, ok := v.(map[string]interface{})
	if !ok {
		panic("cannot get map from decoded file")
	}
	infoM, ok := metaInfo["info"].(map[string]interface{})
	if !ok {
		panic("cannot get map info from metainfo")
	}
	return fmt.Sprintf("%s", metaInfo["announce"]), infoM
}

func extractPiece(v interface{}) []string {
	p, ok := v.(string)
	pieces := make([]string, 0, 10)
	if !ok {
		panic("pieces is not a string")
	}
	for i := 0; i < len(p); i += 20 {
		pieces = append(pieces, fmt.Sprintf("%x", p[i:(i+20)]))
	}
	return pieces
}
