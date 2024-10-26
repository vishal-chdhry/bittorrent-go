package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, string, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := range bencodedString {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], bencodedString[firstColonIndex+1+length:], nil
	} else if bencodedString[0] == 'i' {
		endOfInt := -1
		for i := range bencodedString {
			if bencodedString[i] == 'e' {
				endOfInt = i
				break
			}
		}
		if endOfInt == -1 {
			return -1, "", fmt.Errorf("invalid integer format")
		}
		num, err := strconv.Atoi(bencodedString[1:endOfInt])
		if err != nil {
			return -1, "", err
		}
		return num, bencodedString[endOfInt+1:], nil
	} else if bencodedString[0] == 'l' {
		bencodedString = bencodedString[1:]
		list := []interface{}{}
		for bencodedString[0] != 'e' {
			v, rest, err := decodeBencode(bencodedString)
			if err != nil {
				return -1, "", err
			}
			list = append(list, v)
			bencodedString = rest
		}
		return list, strings.TrimPrefix(bencodedString, "e"), nil
	} else {
		return "", "", fmt.Errorf("Only strings are supported at the moment: %s", bencodedString)
	}

}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
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
