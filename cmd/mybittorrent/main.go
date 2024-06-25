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
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if rune(bencodedString[0]) == 'i' {
		l := len(bencodedString)
		start := 1
		i, err := strconv.Atoi(bencodedString[start:(l - 1)])
		if err != nil {
			return 0, err
		}
		return i, nil
	} else {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func init() {
	if level, err := strconv.ParseBool(os.Getenv("LOG")); err != nil || !level {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	slog.Info("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
