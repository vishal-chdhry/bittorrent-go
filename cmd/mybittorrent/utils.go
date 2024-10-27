package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

func genPeerId() string {
	barray := make([]byte, 10)
	rand.Read(barray)
	return hex.EncodeToString(barray)
}

func writeToDisk(fileName string, fileData []byte) error {
	fo, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = fo.Write(fileData)
	if err != nil {
		return err
	}
	return fo.Close()
}
