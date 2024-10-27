package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"unicode"
)

type bdecoder struct {
	*bufio.Reader
}

func decodeFromBytes(b []byte) (interface{}, error) {
	buf := bytes.NewBuffer(b)
	bd := bdecoder{bufio.NewReader(buf)}
	decoded, err := bd.decode()
	if err != nil {
		return nil, err
	}

	return decoded, err
}

func decodeFromFile(fileName string) (interface{}, error) {
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return decodeFromBytes(b)
}

func (b *bdecoder) decode() (interface{}, error) {
	c, err := b.Peek(1)
	if err != nil {
		return nil, err
	}
	first := c[0]
	switch {
	case unicode.IsDigit(rune(first)):
		return b.decodeString()
	case first == 'i':
		return b.decodeInt()
	case first == 'l':
		return b.decodeList()
	case first == 'd':
		return b.decodeDict()
	default:
		return nil, fmt.Errorf("unsupported type in string or invalid format")
	}
}

func (b *bdecoder) decodeString() (string, error) {
	num, err := b.ReadString(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(num[:len(num)-1])
	if err != nil {
		return "", err
	}
	str := make([]byte, length)
	n, err := b.Read(str)
	if err != nil && err != io.EOF {
		return "", err
	}

	if n != length {
		return "", fmt.Errorf("malformed string")
	}

	return string(str), nil
}

func (b *bdecoder) decodeInt() (int, error) {
	token, err := b.ReadString('e')
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(token[1 : len(token)-1])
}

func (b *bdecoder) decodeList() ([]interface{}, error) {
	b.ReadByte()
	list := make([]interface{}, 0)
	for {
		if c, err := b.Peek(1); err != nil {
			return list, err
		} else if c[0] == 'e' {
			b.ReadByte()
			break
		}

		if val, err := b.decode(); err != nil {
			return list, err
		} else {
			list = append(list, val)
		}
	}
	return list, nil
}

func (b *bdecoder) decodeDict() (map[string]interface{}, error) {
	b.ReadByte()
	dict := make(map[string]interface{})
	for {
		if c, err := b.Peek(1); err != nil {
			return dict, err
		} else if c[0] == 'e' {
			b.ReadByte()
			break
		}

		var key string
		var val interface{}
		var err error
		if key, err = b.decodeString(); err != nil {
			return dict, err
		}
		if val, err = b.decode(); err != nil {
			return dict, err
		}

		dict[key] = val
	}
	return dict, nil
}

type bencoder struct {
	*bytes.Buffer
}

func (b *bencoder) encode(val interface{}) error {
	switch v := val.(type) {
	case string:
		b.WriteString(fmt.Sprintf("%d:%s", len(v), v))
		return nil
	case int:
		b.WriteString(fmt.Sprintf("i%de", v))
		return nil
	case []interface{}:
		b.WriteByte('l')
		for _, el := range v {
			if err := b.encode(el); err != nil {
				return err
			}
		}
		b.WriteByte('e')
		return nil
	case map[string]interface{}:
		b.WriteByte('d')
		m := make([]string, 0, len(v))
		for k := range v {
			m = append(m, k)
		}

		sort.Strings(m)

		for _, key := range m {
			b.encode(key)
			b.encode(v[key])
		}
		b.WriteByte('e')
		return nil
	default:
		return fmt.Errorf("unsupported type in encoder")
	}
}
