package main

import (
	"io"
	"net/http"
)

func fetchPeersFromTorrentUrl(requestUrl string) ([]string, error) {
	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	decoded, err := decodeFromBytes(body)
	if err != nil {
		return nil, err
	}

	peers := decoded.(map[string]interface{})["peers"].(string)
	return parsePeerIPV4s([]byte(peers)), nil
}
