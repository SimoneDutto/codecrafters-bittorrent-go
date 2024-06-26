package main

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/jackpal/bencode-go"
)

func TestBencodeWithLib(t *testing.T) {
	file, err := os.Open("../../sample.torrent")
	if err != nil {
		panic(err)
	}
	decodedLib, _ := bencode.Decode(file)
	fileT, err := os.ReadFile("../../sample.torrent")
	if err != nil {
		panic(err)
	}
	decodeMine, _ := decodeBencode(string(fileT), []interface{}{}, 0)

	if diff := deep.Equal(decodeMine[0], decodedLib); diff != nil {
		t.Error(diff)
	}
}

func TestBencodeToSha(t *testing.T) {
	file, err := os.Open("../../sample.torrent")
	if err != nil {
		panic(err)
	}
	decodedLib, _ := bencode.Decode(file)
	shaW := sha1.New()
	metaInfo, ok := decodedLib.(map[string]interface{})
	if !ok {
		panic("cannot get map from decoded file")
	}
	infoM, ok := metaInfo["info"].(map[string]interface{})
	if !ok {
		panic("cannot get map info from metainfo")
	}
	bencode.Marshal(shaW, infoM)
	shaS := hex.EncodeToString(shaW.Sum(nil))
	if shaS != "d69f91e6b2ae4c542468d1073a71d4ea13879a7f" {
		t.Fatalf("sha %s different\n", shaS)
	}
}
