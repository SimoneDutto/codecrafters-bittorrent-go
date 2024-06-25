package main

import (
	// Uncomment this line to pass the first stage
	// "encoding/json"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string, elems []interface{}, start int) ([]interface{}, int) {
	slog.Info(fmt.Sprintf("start: %d", start))
	slog.Info(fmt.Sprintf("elems: %#v", elems))
	if len(bencodedString) == start {
		return elems, start
	}
	slog.Info(fmt.Sprintf("start char %c", bencodedString[start]))
	if rune(bencodedString[start]) == 'l' {
		slog.Info("list detected")
		encodedL, end := decodeBencode(bencodedString, []interface{}{}, start+1)
		elems = append(elems, encodedL)
		return decodeBencode(bencodedString, elems, start+end)
	} else if rune(bencodedString[start]) == 'e' {
		slog.Info("detected end")
		return decodeBencode(bencodedString, elems, start+1)
	} else if unicode.IsDigit(rune(bencodedString[start])) {
		slog.Info("string detected")
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
			panic("cannot convert string len")
		}

		elem := bencodedString[firstColonIndex+1 : firstColonIndex+1+length]
		elems = append(elems, elem)
		return decodeBencode(bencodedString, elems, firstColonIndex+1+length)
	} else if rune(bencodedString[start]) == 'i' {
		slog.Info("integer detected")
		l := start + peekUntil(bencodedString, start, 'e')
		startI := start + 1
		i, err := strconv.Atoi(bencodedString[startI:l])
		if err != nil {
			panic("cannot convert string to int")
		}
		elem := i
		elems = append(elems, elem)
		return decodeBencode(bencodedString, elems, start+l)
	} else {
		panic("not supported type")
	}
}

func init() {
	if level, err := strconv.ParseBool(os.Getenv("LOG")); err != nil || !level {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}

func main() {
	slog.Info("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, end := decodeBencode(bencodedValue, []interface{}{}, 0)
		slog.Info(fmt.Sprintf("Number of elements decoded: %d", len(decoded)))
		slog.Info(fmt.Sprintf("Number of characters: %d", end))
		jsonOutput, _ := json.Marshal(decoded[0])
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func peekUntil(s string, start int, charEnd rune) int {
	for i := start; i < len(s); i++ {
		if s[i] == byte(charEnd) {
			return i - 1
		}
	}
	panic("cannot find peek char")
}
