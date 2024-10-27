package main

import (
	"encoding/binary"
	"net"
)

func receiveMsgInfo(connection net.Conn) (uint32, byte, error) {
	messageTypeBuffer := make([]byte, 5)
	_, err := connection.Read(messageTypeBuffer)
	if err != nil {
		return 0, 0, err
	}
	length := binary.BigEndian.Uint32(messageTypeBuffer[:4])
	messageType := messageTypeBuffer[4]
	return length - 1, messageType, nil
}

func buildMessage(msgType byte, msg []byte) []byte {
	var length uint32 = uint32(len(msg)) + 1

	payload := make([]byte, 0, 4+1+len(msg))
	payload = binary.BigEndian.AppendUint32(payload, uint32(length))
	payload = append(payload, msgType)
	payload = append(payload, msg...)
	return payload
}

func buildDownloadRequest(pieceIndex, blockOffset, length int) []byte {
	payload := make([]byte, 0, 12)
	payload = binary.BigEndian.AppendUint32(payload, uint32(pieceIndex))
	payload = binary.BigEndian.AppendUint32(payload, uint32(blockOffset))
	payload = binary.BigEndian.AppendUint32(payload, uint32(length))
	return payload
}
