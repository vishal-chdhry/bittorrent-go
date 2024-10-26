package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string, start int) (interface{}, int, error) {
	if start == len(bencodedString) {
		return nil, -1, io.ErrUnexpectedEOF
	}
	c := bencodedString[start]
	switch {
	case c == 'i':
		val, next, err := decodeInt(bencodedString, start)
		if err != nil {
			return -1, -1, err
		}
		return val, next, nil
	case c == 'l':
		list, next, err := decodeList(bencodedString, start)
		if err != nil {
			return list, -1, err
		}
		return list, next, nil
	case c == 'd':
		dict, next, err := decodeDict(bencodedString, start)
		if err != nil {
			return dict, -1, err
		}
		return dict, next, nil
	case c >= '0' && c <= '9':
		str, next, err := decodeString(bencodedString, start)
		if err != nil {
			return "", -1, err
		}
		return str, next, nil
	default:
		return nil, -1, fmt.Errorf("unsupported type provided: %s", bencodedString)
	}
}

func decodeString(bencodedString string, start int) (string, int, error) {
	firstColonIndex := -1

	for i := start; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}
	if firstColonIndex == -1 {
		return "", -1, fmt.Errorf("invalid string format, %s", bencodedString[start:])
	}

	lengthStr := bencodedString[start:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", -1, err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], firstColonIndex + 1 + length, nil
}

func decodeInt(bencodedString string, start int) (int, int, error) {
	endOfInt := -1
	for i := start; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			endOfInt = i
			break
		}
	}
	if endOfInt == -1 {
		return -1, -1, fmt.Errorf("invalid integer format, %s", bencodedString[start:])
	}
	num, err := strconv.Atoi(bencodedString[start+1 : endOfInt])
	if err != nil {
		return -1, -1, err
	}
	return num, endOfInt + 1, nil
}

func decodeList(bencodedString string, start int) ([]interface{}, int, error) {
	list := []interface{}{}
	beg := start + 1
	for bencodedString[beg] != 'e' {
		v, next, err := decodeBencode(bencodedString, beg)
		if err != nil {
			return list, -1, err
		}
		list = append(list, v)
		beg = next
	}
	return list, beg + 1, nil
}

func decodeDict(bencodedString string, start int) (map[string]interface{}, int, error) {
	dict := map[string]interface{}{}
	beg := start + 1
	for bencodedString[beg] != 'e' {
		key, next, err := decodeString(bencodedString, beg)
		if err != nil {
			return dict, -1, err
		}
		beg = next
		value, next, err := decodeBencode(bencodedString, beg)
		if err != nil {
			return dict, -1, err
		}
		beg = next
		dict[key] = value
	}

	// sort the dictionary
	// keys := make([]string, 0, len(dict))
	// sort.Strings(keys)
	// sortedDict :=

	return dict, beg + 1, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue, 0)
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
