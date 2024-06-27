package main

import (
	// Uncomment this line to pass the first stage
	// "encoding/json"

	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"unicode"

	"github.com/jackpal/bencode-go"
	// Available if you need it!
	// Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string, elems []interface{}, start int) ([]interface{}, int) {
	slog.Debug(fmt.Sprintf("start: %d", start))
	slog.Debug(fmt.Sprintf("elems: %#v", elems))
	if len(bencodedString) == start {
		return elems, start
	}
	slog.Debug(fmt.Sprintf("start char %s", bencodedString[start:]))
	if rune(bencodedString[start]) == 'l' {
		slog.Debug("list detected")
		encodedL, end := decodeBencode(bencodedString, []interface{}{}, start+1)
		slog.Debug(fmt.Sprintf("elems from list: %#v", encodedL))
		elems = append(elems, encodedL)
		return decodeBencode(bencodedString, elems, end)
	} else if rune(bencodedString[start]) == 'e' {
		slog.Debug("detected end")
		return elems, start + 1
	} else if rune(bencodedString[start]) == 'd' {
		encodedL, end := decodeBencode(bencodedString, []interface{}{}, start+1)
		m := make(map[string]interface{})
		for i := 0; i < len(encodedL); i += 2 {
			val, ok := encodedL[i].(string)
			if !ok {
				panic("dictionary key is not string")
			}
			m[val] = encodedL[i+1]
		}
		slog.Debug(fmt.Sprintf("elems from map: %#v", m))
		elems = append(elems, m)
		return decodeBencode(bencodedString, elems, end)
	} else if unicode.IsDigit(rune(bencodedString[start])) {
		slog.Debug("string detected")
		var firstColonIndex int
		for i := start; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}
		lengthStr := bencodedString[start:firstColonIndex]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			panic("cannot convert string len array")
		}

		elem := bencodedString[firstColonIndex+1 : firstColonIndex+1+length]
		elems = append(elems, elem)
		return decodeBencode(bencodedString, elems, firstColonIndex+1+length)
	} else if rune(bencodedString[start]) == 'i' {
		slog.Debug("integer detected")
		l := peekUntil(bencodedString, start, 'e')
		startI := start + 1
		i, err := strconv.Atoi(bencodedString[startI:l])
		if err != nil {
			panic("cannot convert string to int")
		}
		elem := int64(i)
		elems = append(elems, elem)
		return decodeBencode(bencodedString, elems, l+1)
	} else {
		panic("not supported type")
	}
}

// TODO: remove lib and implement it yourself
func bencodeBencode(val interface{}) string {
	b := new(bytes.Buffer)
	err := bencode.Marshal(b, val)
	if err != nil {
		panic(err)
	}
	return b.String()
}

func init() {
	// if level, err := strconv.ParseBool(os.Getenv("LOG")); err != nil || !level {
	// 	slog.SetLogLoggerLevel(slog.LevelError)
	// }

}

func main() {
	slog.Info("Logs from your program will appear here!")
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, end := decodeBencode(bencodedValue, []interface{}{}, 0)
		slog.Info(fmt.Sprintf("Number of elements decoded: %d", len(decoded)))
		slog.Info(fmt.Sprintf("Number of characters: %d", end))
		slog.Info(fmt.Sprintf("elems from decode: %#v", decoded))
		for _, d := range decoded {
			jsonOutput, err := json.Marshal(d)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(jsonOutput))
		}
	} else if command == "info" {
		file := os.Args[2]
		bF, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		decoded, _ := decodeBencode(string(bF), []interface{}{}, 0)
		announce, infoM := extractInfo(decoded[0])
		fmt.Printf("Tracker URL: %s\n", announce)
		fmt.Printf("Length: %d\n", infoM["length"])
		fmt.Printf("Info Hash: %s\n", hex.EncodeToString(calcSha1([]byte(bencodeBencode(infoM)))))
		fmt.Printf("Piece Length: %d\n", infoM["piece length"])
		fmt.Println("Piece Hashes:")
		pieces := extractPiece(infoM["pieces"])
		for _, p := range pieces {
			fmt.Printf("%x\n", p)
		}
	} else if command == "peers" {
		file := os.Args[2]
		bF, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		decoded, _ := decodeBencode(string(bF), []interface{}{}, 0)
		announce, infoM := extractInfo(decoded[0])
		hashInfo := calcSha1([]byte(bencodeBencode(infoM)))
		peers := getPeers(announce, hashInfo, infoM["piece length"])
		slog.Info(fmt.Sprintf("extracted %d peers\n", len(peers)))
		for _, p := range peers {
			fmt.Println(p)
		}
	} else if command == "handshake" {
		file := os.Args[2]
		endpoint := os.Args[3]
		bF, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		decoded, _ := decodeBencode(string(bF), []interface{}{}, 0)
		_, infoM := extractInfo(decoded[0])
		hashInfo := calcSha1([]byte(bencodeBencode(infoM)))
		conn, err := net.Dial("tcp", endpoint)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		res := sendHandskake(conn, hashInfo)
		fmt.Printf("Peer ID: %s\n", hex.EncodeToString(res[len(res)-20:]))
	} else if command == "download_piece" {
		file := os.Args[4]
		fileO := os.Args[3]
		n, _ := strconv.Atoi(os.Args[5])
		bF, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		decoded, _ := decodeBencode(string(bF), []interface{}{}, 0)
		announce, infoM := extractInfo(decoded[0])
		hashInfo := calcSha1([]byte(bencodeBencode(infoM)))
		slog.Info(fmt.Sprintf("InfoM: %#v\n", infoM))
		peers := getPeers(announce, hashInfo, infoM["piece length"])
		endpoint := peers[0]
		conn, err := net.Dial("tcp", endpoint)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		res := sendHandskake(conn, hashInfo)
		slog.Info(fmt.Sprintf("Handshake: %#v\n", res))
		unchoke(conn)
		f, err := os.Create(fileO)
		if err != nil {
			panic(err)
		}
		piece := downloadPiece(conn, fileO, uint32(n), uint32(infoM["piece length"].(int64)), uint32(infoM["length"].(int64)))
		f.Write(piece)
	} else if command == "download" {
		file := os.Args[4]
		fileO := os.Args[3]
		bF, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		decoded, _ := decodeBencode(string(bF), []interface{}{}, 0)
		announce, infoM := extractInfo(decoded[0])
		hashInfo := calcSha1([]byte(bencodeBencode(infoM)))
		slog.Info(fmt.Sprintf("InfoM: %#v\n", infoM))
		peers := getPeers(announce, hashInfo, infoM["piece length"])
		endpoint := peers[0]
		conn, err := net.Dial("tcp", endpoint)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		res := sendHandskake(conn, hashInfo)
		slog.Info(fmt.Sprintf("Handshake: %#v\n", res))
		unchoke(conn)
		f, err := os.Create(fileO)
		if err != nil {
			panic(err)
		}
		for i := range extractPiece(infoM["pieces"]) {
			piece := downloadPiece(conn, fileO, uint32(i), uint32(infoM["piece length"].(int64)), uint32(infoM["length"].(int64)))
			f.Write(piece)
		}

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
