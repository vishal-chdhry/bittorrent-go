package main

import (
	"bytes"
	"fmt"
	"net"
)

func getMagnetExtensionPayload(conn net.Conn) (interface{}, error) {
	// accept bitfield msg
	msglength, msgType, err := receiveMsgInfo(conn)
	if err != nil {
		return nil, err
	} else if msgType != 5 {
		return nil, fmt.Errorf("expected msg type: bitfiled, received %d", msgType)
	}

	// assume all peer have all the files so ignore the message
	flushBytesFromConn(conn, msglength)

	// send extension handshake msg
	handshakeMsg := []byte{0}
	hndshkePayload, err := buildExtensionHandshakeMessage()
	if err != nil {
		return nil, err
	}

	handshakeMsg = append(handshakeMsg, hndshkePayload...)
	_, err = conn.Write(buildMessage(20, handshakeMsg))
	if err != nil {
		return nil, err
	}

	// receive extension handshake msg
	msglength, msgType, err = receiveMsgInfo(conn)
	if err != nil {
		return nil, err
	} else if msgType != 20 {
		return nil, fmt.Errorf("expected msg type: extension, received %d", msgType)
	}

	msg := make([]byte, msglength)
	_, err = conn.Read(msg)
	if err != nil {
		return nil, err
	}
	payloadMap := msg[1:] // first byte is the extension message id
	decoded, err := decodeFromBytes(payloadMap)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", decoded)
	return decoded, nil
}

func buildExtensionHandshakeMessage() ([]byte, error) {
	payloadMap := make(map[string]interface{})
	payloadMap["m"] = map[string]interface{}{
		"ut_metadata": 69,
	}

	buf := bytes.Buffer{}
	be := bencoder{&buf}
	err := be.encode(payloadMap)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
