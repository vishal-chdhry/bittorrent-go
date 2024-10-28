package main

import (
	"fmt"
	"net"
)

func connectWithPeer(peerAddress string, clientId string, infoHash []byte, extension []byte) (net.Conn, []byte, error) {
	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return nil, nil, err
	}
	pstrlen := byte(19) // The length of the string "BitTorrent protocol"
	pstr := []byte("BitTorrent protocol")
	reserved := make([]byte, 8) // Eight zeros
	if len(extension) != 0 {
		reserved = extension
	}
	handshake := append([]byte{pstrlen}, pstr...)
	handshake = append(handshake, reserved...)
	handshake = append(handshake, infoHash...)
	handshake = append(handshake, []byte(clientId)...)
	_, err = conn.Write([]byte(handshake))
	handshakebuffer := make([]byte, 1+19+8+20+20)

	_, err = conn.Read(handshakebuffer)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil, err
	}

	return conn, handshakebuffer[1+19+8+20:], nil
}

func initiateRcvRequest(conn net.Conn) error {
	// assume all peer have all the files
	msglength, msgType, err := receiveMsgInfo(conn)
	if err != nil {
		return err
	} else if msgType != 5 {
		return fmt.Errorf("expected msg type: bitfiled, received %d", msgType)
	}

	flushBytesFromConn(conn, msglength)

	// send interest message
	if _, err := conn.Write(buildMessage(2, nil)); err != nil {
		return err
	}

	// wait for unchoke message
	if msgLen, msgType, err := receiveMsgInfo(conn); err != nil {
		return err
	} else if msgType != 1 || msgLen != 0 {
		return fmt.Errorf("expected msg type: unchoke, received %d, length %d", msgType, msgLen)
	}

	return nil
}

func flushBytesFromConn(conn net.Conn, n uint32) error {
	_, err := conn.Read(make([]byte, n))
	return err
}

func enableMagnetExtension() []byte {
	a := make([]byte, 8)
	a[5] = 16
	return a
}
