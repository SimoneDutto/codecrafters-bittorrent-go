package main

import (
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
