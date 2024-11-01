package main

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

func parseMagentFromString(m string) (map[string]string, error) {
	if !strings.HasPrefix(m, "magnet:?") {
		return nil, fmt.Errorf("invalid magnet link")
	}
	var err error
	m, err = url.QueryUnescape(m)
	if err != nil {
		return nil, err
	}

	val := make(map[string]string)
	m = strings.TrimPrefix(m, "magnet:?")
	parts := strings.Split(m, "&")

	for _, el := range parts {
		parts := strings.Split(el, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid magnet link")
		}
		val[parts[0]] = parts[1]
	}
	stringCodedInfoHash := strings.TrimPrefix(val["xt"], "urn:btih:")
	h, err := hex.DecodeString(stringCodedInfoHash)
	if err != nil {
		return nil, err
	}
	val["xt"] = string(h)
	return val, nil
}
