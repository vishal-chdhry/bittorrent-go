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
	handshakeMsg := []byte{0} // extension message id
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

func getMagnetRequestMetadata(conn net.Conn, extId byte) (interface{}, error) {
	reqBytes := []byte{extId}
	p, err := buildMagnetRequestPayload()
	if err != nil {
		fmt.Println("youJHK")
		return nil, err
	}
	reqBytes = append(reqBytes, p...)
	// write request payload
	_, err = conn.Write(buildMessage(20, reqBytes))
	if err != nil {
		return nil, err
	}

	fmt.Println("why nil")
	// receive extension request response
	msglength, msgType, err := receiveMsgInfo(conn)
	if err != nil {
		return nil, err
	} else if msgType != 20 {
		return nil, fmt.Errorf("expected msg type: extension, received %d", msgType)
	}

	fmt.Println(msglength)

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
	fmt.Println(decoded)
	return decoded, nil
}

func buildMagnetRequestPayload() ([]byte, error) {
	payloadMap := make(map[string]interface{})
	payloadMap["msg_type"] = 0
	payloadMap["piece"] = 0

	buf := bytes.Buffer{}
	be := bencoder{&buf}
	err := be.encode(payloadMap)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return buf.Bytes(), nil
}
