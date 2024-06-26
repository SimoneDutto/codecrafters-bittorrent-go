package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
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
		pieces = append(pieces, p[i:(i+20)])
	}
	return pieces
}

func prepareRequest(hash string, length int64) string {
	params := url.Values{}
	params.Add("info_hash", hash)
	params.Add("peer_id", "01238183173107890890")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprint(length))
	params.Add("compact", "1")
	return params.Encode()
}
