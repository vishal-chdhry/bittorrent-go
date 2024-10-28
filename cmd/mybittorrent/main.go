package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

func main() {
	if len(os.Args) < 3 {
		fmt.Println("invalid arguments provided, there should be atleast three arguments")
		return
	}
	command := os.Args[1]
	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		decoded, err := decodeFromBytes([]byte(bencodedValue))
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
		return
	case "info":
		torrentInfo, err := getTorrentInfoFromFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Tracker URL:", torrentInfo.TrackerURL)
		fmt.Println("Length:", torrentInfo.FileLength)
		fmt.Printf("Info Hash: %x\n", torrentInfo.InfoHash)
		fmt.Printf("Piece Length: %d\n", torrentInfo.PieceLength)
		fmt.Printf("Piece Hashes:\n")
		for _, v := range torrentInfo.PieceHashes {
			fmt.Println(v)
		}
		return
	case "peers":
		torrentInfo, err := getTorrentInfoFromFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		u := getRequestUrlFromTorrentInfo(torrentInfo)
		peerUrls, err := fetchPeersFromTorrentUrl(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, v := range peerUrls {
			fmt.Println(v)
		}
		return
	case "handshake":
		torrentInfo, err := getTorrentInfoFromFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		u := getRequestUrlFromTorrentInfo(torrentInfo)
		_, err = fetchPeersFromTorrentUrl(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		conn, peerId, err := connectWithPeer(os.Args[3], genPeerId(), torrentInfo)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()
		fmt.Printf("Peer ID: %x\n", peerId)
		return
	case "download_piece":
		torrentInfo, err := getTorrentInfoFromFile(os.Args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
		u := getRequestUrlFromTorrentInfo(torrentInfo)
		peerUrls, err := fetchPeersFromTorrentUrl(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		clientId := genPeerId()
		conn, _, err := connectWithPeer(peerUrls[0], clientId, torrentInfo)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()

		if err := initiateRcvRequest(conn); err != nil {
			fmt.Println("Error:", err)
			return
		}
		index, _ := strconv.ParseInt(os.Args[5], 10, 32)
		i := int(index)
		fileData, err := downloadPiece(conn, torrentInfo.PieceLength, i, torrentInfo.FileLength)
		if err != nil {
			return
		}
		err = writeToDisk(os.Args[3], fileData)
		if err != nil {
			fmt.Println(err)
			return
		}

		return
	case "download":
		torrentInfo, err := getTorrentInfoFromFile(os.Args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
		u := getRequestUrlFromTorrentInfo(torrentInfo)
		peerUrls, err := fetchPeersFromTorrentUrl(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		clientId := genPeerId()
		connMap := make([]net.Conn, 0, len(peerUrls))
		for _, peer := range peerUrls {
			conn, _, err := connectWithPeer(peer, clientId, torrentInfo)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer conn.Close()

			if err := initiateRcvRequest(conn); err != nil {
				fmt.Println("Error:", err)
				return
			}
			connMap = append(connMap, conn)
		}

		fileData := make([]byte, 0)
		fileMap := make(map[int][]byte)
		wq := createWorkQueue(torrentInfo)
		workers := createWorkers(torrentInfo, connMap, fileMap)
		wPool := newWorkerPool(wq, workers...)
		wPool.start()
		for i := range torrentInfo.PieceHashes {
			fileData = append(fileData, fileMap[i]...)
		}
		// fmt.Println(len(fileData))
		err = writeToDisk(os.Args[3], fileData)
		if err != nil {
			fmt.Println(err)
			return
		}

		return
	case "magnet_parse":
		if len(os.Args) != 3 {
			fmt.Println("usage: ./your_bittorrent.sh magnet_parse <magnet-link>")
			os.Exit(1)
		}

		magnetLink := os.Args[2]
		mag, err := parseMagentFromString(magnetLink)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Tracker URL:", mag["tr"])
		fmt.Println("Info Hash:", strings.TrimPrefix(mag["xt"], "urn:btih:"))
	default:
		fmt.Println("unsupported command", command)
		return
	}
}
