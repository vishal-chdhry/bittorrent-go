package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"unicode"
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

type bdecoder struct {
	*bufio.Reader
}

func (b *bdecoder) decode() (interface{}, error) {
	c, err := b.Peek(1)
	if err != nil {
		return nil, err
	}
	first := c[0]
	switch {
	case unicode.IsDigit(rune(first)):
		return b.decodeString()
	case first == 'i':
		return b.decodeInt()
	case first == 'l':
		return b.decodeList()
	case first == 'd':
		return b.decodeDict()
	default:
		return nil, fmt.Errorf("unsupported type in string or invalid format")
	}
}

func (b *bdecoder) decodeString() (string, error) {
	num, err := b.ReadString(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(num[:len(num)-1])
	if err != nil {
		return "", err
	}
	str := make([]byte, length)
	n, err := b.Read(str)
	if err != nil && err != io.EOF {
		return "", err
	}

	if n != length {
		return "", fmt.Errorf("malformed string")
	}

	return string(str), nil
}

func (b *bdecoder) decodeInt() (int, error) {
	token, err := b.ReadString('e')
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(token[1 : len(token)-1])
}

func (b *bdecoder) decodeList() ([]interface{}, error) {
	b.ReadByte()
	list := make([]interface{}, 0)
	for {
		if c, err := b.Peek(1); err != nil {
			return list, err
		} else if c[0] == 'e' {
			b.ReadByte()
			break
		}

		if val, err := b.decode(); err != nil {
			return list, err
		} else {
			list = append(list, val)
		}
	}
	return list, nil
}

func (b *bdecoder) decodeDict() (map[string]interface{}, error) {
	b.ReadByte()
	dict := make(map[string]interface{})
	for {
		if c, err := b.Peek(1); err != nil {
			return dict, err
		} else if c[0] == 'e' {
			b.ReadByte()
			break
		}

		var key string
		var val interface{}
		var err error
		if key, err = b.decodeString(); err != nil {
			return dict, err
		}
		if val, err = b.decode(); err != nil {
			return dict, err
		}

		dict[key] = val
	}
	return dict, nil
}

type bencoder struct {
	*bytes.Buffer
}

func (b *bencoder) encode(val interface{}) error {
	switch v := val.(type) {
	case string:
		b.WriteString(fmt.Sprintf("%d:%s", len(v), v))
		return nil
	case int:
		b.WriteString(fmt.Sprintf("i%de", v))
		return nil
	case []interface{}:
		b.WriteByte('l')
		for _, el := range v {
			if err := b.encode(el); err != nil {
				return err
			}
		}
		b.WriteByte('e')
		return nil
	case map[string]interface{}:
		b.WriteByte('d')
		m := make([]string, 0, len(v))
		for k := range v {
			m = append(m, k)
		}

		sort.Strings(m)

		for _, key := range m {
			b.encode(key)
			b.encode(v[key])
		}
		b.WriteByte('e')
		return nil
	default:
		return fmt.Errorf("unsupported type in encoder")
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("invalid arguments provided, there should be atleast three arguments")
		return
	}
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		buf := bytes.NewBuffer([]byte(bencodedValue))
		bd := bdecoder{bufio.NewReader(buf)}
		decoded, err := bd.decode()
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fileName := os.Args[2]
		if command == "download_piece" {
			fileName = os.Args[4]
		}

		f, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		bd := bdecoder{bufio.NewReader(f)}
		decoded, err := bd.decode()
		if err != nil {
			fmt.Println(err)
			return
		}
		infoMap := decoded.(map[string]interface{})["info"].(map[string]interface{})
		buf := bytes.Buffer{}
		be := bencoder{&buf}
		err = be.encode(infoMap)
		if err != nil {
			fmt.Println(err)
			return
		}
		h := sha1.New()
		io.Copy(h, &buf)

		sum := h.Sum(nil)

		trackerUrl := decoded.(map[string]interface{})["announce"].(string)
		length := infoMap["length"].(int)
		infoHash := hex.EncodeToString(sum)
		pieceLength := infoMap["piece length"].(int)
		pieces := make([]string, 0)
		b := bytes.NewBuffer([]byte(infoMap["pieces"].(string)))
		for b.Len() != 0 {
			hash := make([]byte, 20)
			_, err := b.Read(hash)
			if err != nil {
				fmt.Println(err)
				return
			}
			pieces = append(pieces, hex.EncodeToString(hash))
		}

		barray := make([]byte, 10)
		rand.Read(barray)
		peerId := hex.EncodeToString(barray)

		if command == "info" {
			fmt.Println("Tracker URL:", trackerUrl)
			fmt.Println("Length:", length)
			fmt.Printf("Info Hash: %s\n", infoHash)
			fmt.Printf("Piece Length: %d\n", pieceLength)
			fmt.Printf("Piece Hashes:\n")
			for _, v := range pieces {
				fmt.Println(v)
			}
		} else if command == "peers" || command == "handshake" || command == "download_piece" {
			val := url.Values{}
			val.Add("peer_id", peerId)
			val.Add("port", "6881")
			val.Add("uploaded", "0")
			val.Add("downloaded", "0")
			val.Add("left", fmt.Sprint(length))
			val.Add("compact", "1")
			u := trackerUrl + "?" + val.Encode() + "&info_hash=" + url.QueryEscape(string(sum))
			resp, err := http.Get(u)
			if err != nil {
				fmt.Println(err)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			buf := bytes.NewBuffer([]byte(body))
			bd := bdecoder{bufio.NewReader(buf)}
			decoded, err := bd.decode()
			if err != nil {
				fmt.Println(err)
				return
			}

			peers := decoded.(map[string]interface{})["peers"].(string)
			peeripv4s := parsePeerIPV4s([]byte(peers))
			if command == "peers" {
				for _, v := range peeripv4s {
					fmt.Println(v)
				}
			} else {
				peerAddress := peeripv4s[0]
				conn, err := net.Dial("tcp", peerAddress)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer conn.Close()
				pstrlen := byte(19) // The length of the string "BitTorrent protocol"
				pstr := []byte("BitTorrent protocol")
				reserved := make([]byte, 8) // Eight zeros
				handshake := append([]byte{pstrlen}, pstr...)
				handshake = append(handshake, reserved...)
				handshake = append(handshake, sum...)
				handshake = append(handshake, []byte(peerId)...)
				_, err = conn.Write([]byte(handshake))
				handshakebuffer := make([]byte, 1+19+8+20+20)

				_, err = conn.Read(handshakebuffer)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				if command == "handshake" {
					recieverPeerId := handshakebuffer[1+19+8+20:]
					fmt.Printf("Peer ID: %x\n", recieverPeerId)
				} else {
					// all peers have all package so ignore
					_, _, err := getMessageInfo(conn)
					if err != nil {
						fmt.Println("Error:", err)
						return
					}
					_, err = conn.Read(make([]byte, 1))
					if _, err = conn.Write(createMessage(2, nil)); err != nil {
						fmt.Println("Error:", err)
						return
					}

					// unchoke message
					if length, msgType, err := getMessageInfo(conn); err != nil {
						fmt.Println("Error:", err)
						return
					} else if msgType != 1 || length != 0 {
						fmt.Println("wrong message, should have been unchoke")
					}
					fileData := make([]byte, 0)
					// make piece requests
					// for i := range pieces {
					i, _ := strconv.ParseInt(os.Args[5], 10, 32)
					blockSize := int(math.Pow(2, 14))
					numBlocks := int(math.Ceil(float64(pieceLength) / float64(blockSize)))
					for j := 0; j < numBlocks; j++ {
						blockLength, eof := calculateBlockLength(length, pieceLength, blockSize, int(i), j)
						if _, err = conn.Write(createMessage(6, createRequestPayload(int(i), j*blockSize, blockLength))); err != nil {
							fmt.Println("Error:", err)
							return
						}

						length, msgType, err := getMessageInfo(conn)
						if err != nil {
							fmt.Println("Error:", err)
							return
						} else if msgType != 7 {
							fmt.Println("expected a piece msg")
							return
						}
						fmt.Println(length, msgType)
						_, err = conn.Read(make([]byte, 8))
						bytesRead := uint32(8)
						msg := make([]byte, length)
						for bytesRead != length {
							n, err := conn.Read(msg)
							if err != nil {
								fmt.Println("Error:", err)
								return
							}
							bytesRead += uint32(n)
							fileData = append(fileData, msg[:n]...)
						}
						if eof {
							break
						}
						fmt.Println(bytesRead)
					}
					shaval := sha1.Sum(fileData)
					fmt.Printf("%x\n", shaval)
					fmt.Println(pieces[i])
					outputFileName := os.Args[3]
					fo, err := os.Create(outputFileName)
					if err != nil {
						fmt.Println(err)
						return
					}
					_, err = fo.Write(fileData)
					if err != nil {
						fmt.Println(err)
						return
					}
					fo.Close()
				}
				// }

			}
		} else {
			fmt.Println("Unknown command: " + command)
			os.Exit(1)
		}
	}
}

func parsePeerIPV4s(ips []byte) []string {
	ipAddrs := make([]string, 0, len(ips)/6)
	for i := 0; i < len(ips); i += 6 {
		ipAddrs = append(ipAddrs, fmt.Sprintf("%d.%d.%d.%d:%d", ips[i], ips[i+1], ips[i+2], ips[i+3], binary.BigEndian.Uint16(ips[i+4:i+6])))
	}
	return ipAddrs
}

func getMessageInfo(connection net.Conn) (uint32, byte, error) {
	messageTypeBuffer := make([]byte, 5)
	_, err := connection.Read(messageTypeBuffer)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, 0, err
	}
	length := binary.BigEndian.Uint32(messageTypeBuffer[:4])
	messageType := messageTypeBuffer[4]
	return length - 1, messageType, nil
}

func createMessage(msgType byte, msg []byte) []byte {
	var length uint32 = uint32(len(msg)) + 1

	payload := make([]byte, 0, 4+1+len(msg))
	payload = binary.BigEndian.AppendUint32(payload, uint32(length))
	payload = append(payload, msgType)
	payload = append(payload, msg...)
	return payload
}

func createRequestPayload(pieceIndex, blockOffset, length int) []byte {
	payload := make([]byte, 0, 12)
	payload = binary.BigEndian.AppendUint32(payload, uint32(pieceIndex))
	payload = binary.BigEndian.AppendUint32(payload, uint32(blockOffset))
	payload = binary.BigEndian.AppendUint32(payload, uint32(length))
	return payload
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
