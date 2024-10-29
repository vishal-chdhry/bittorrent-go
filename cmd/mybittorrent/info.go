package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
)

type torrentInfo struct {
	TrackerURL  string
	Name        string
	FileLength  int
	InfoHash    []byte
	PieceLength int
	PieceHashes []string
}

func getTorrentInfo(decoded interface{}) (*torrentInfo, error) {
	infoMap, ok := decoded.(map[string]interface{})["info"].(map[string]interface{})
	if !ok {
		infoMap = decoded.(map[string]interface{})
	}
	buf := bytes.Buffer{}
	be := bencoder{&buf}
	err := be.encode(infoMap)
	if err != nil {
		return nil, err
	}
	h := sha1.New()
	io.Copy(h, &buf)

	sum := h.Sum(nil)

	trackerUrl, ok := decoded.(map[string]interface{})["announce"].(string)
	if !ok {
		trackerUrl = "~"
	}
	name, ok := infoMap["name"].(string)
	if !ok {
		name = "~"
	}

	length := infoMap["length"].(int)
	pieceLength := infoMap["piece length"].(int)
	pieces := make([]string, 0)
	b := bytes.NewBuffer([]byte(infoMap["pieces"].(string)))
	for b.Len() != 0 {
		hash := make([]byte, 20)
		_, err := b.Read(hash)
		if err != nil {
			return nil, err
		}
		pieces = append(pieces, hex.EncodeToString(hash))
	}

	return &torrentInfo{
		TrackerURL:  trackerUrl,
		Name:        name,
		FileLength:  length,
		InfoHash:    sum,
		PieceLength: pieceLength,
		PieceHashes: pieces,
	}, nil
}

func getTorrentInfoFromFile(filename string) (*torrentInfo, error) {
	decoded, err := decodeFromFile(filename)
	if err != nil {
		return nil, err
	}

	return getTorrentInfo(decoded)
}
