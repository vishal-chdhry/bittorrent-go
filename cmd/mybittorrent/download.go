package main

import (
	"fmt"
	"math"
	"net"
)

func downloadPiece(conn net.Conn, pieceLength, pieceIdx, fileLength int) ([]byte, error) {
	pieceData := make([]byte, 0)
	blockSize := int(math.Pow(2, 14))
	numBlocks := int(math.Ceil(float64(pieceLength) / float64(blockSize)))

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		blockLength, eof := calculateBlockLength(fileLength, pieceLength, blockSize, pieceIdx, blockIdx)
		if _, err := conn.Write(buildMessage(6, buildDownloadRequest(pieceIdx, blockIdx*blockSize, blockLength))); err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}

		length, msgType, err := receiveMsgInfo(conn)
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		} else if msgType != 7 {
			fmt.Println("expected a piece msg")
			return nil, err
		}
		fmt.Println(length, msgType)
		_, err = conn.Read(make([]byte, 8))
		bytesRead := uint32(8)
		msg := make([]byte, length)
		for bytesRead != length {
			n, err := conn.Read(msg)
			if err != nil {
				fmt.Println("Error:", err)
				return nil, err
			}
			bytesRead += uint32(n)
			pieceData = append(pieceData, msg[:n]...)
		}
		fmt.Println(bytesRead)
		if eof {
			break
		}
	}
	return pieceData, nil
}

func calculateBlockLength(totalLength, pieceLength, maxBlockLength, pieceIndex, blockIndex int) (int, bool) {
	numPieces := int(math.Ceil(float64(totalLength) / float64(pieceLength)))
	numBlocks := int(math.Ceil(float64(pieceLength) / float64(maxBlockLength)))
	if pieceIndex >= numPieces || blockIndex >= numBlocks {
		return 0, true
	}

	lastPieceLength := pieceLength - (numPieces*pieceLength - totalLength)
	if pieceIndex == numPieces-1 {
		numBlocks := int(math.Ceil(float64(lastPieceLength) / float64(maxBlockLength)))
		if blockIndex == numBlocks-1 {
			lastBlockLength := lastPieceLength - maxBlockLength*(numBlocks-1)
			return lastBlockLength, true
		}
	}
	return maxBlockLength, false
}
