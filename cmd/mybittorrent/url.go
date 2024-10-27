package main

import (
	"encoding/binary"
	"fmt"
	"net/url"
)

func getRequestUrlFromTorrentInfo(tInfo *torrentInfo) string {
	peerId := genPeerId()
	val := url.Values{}
	val.Add("peer_id", peerId)
	val.Add("port", "6881")
	val.Add("uploaded", "0")
	val.Add("downloaded", "0")
	val.Add("left", fmt.Sprint(tInfo.FileLength))
	val.Add("compact", "1")

	return tInfo.TrackerURL + "?" + val.Encode() + "&info_hash=" + url.QueryEscape(string(tInfo.InfoHash))
}

func parsePeerIPV4s(ips []byte) []string {
	ipAddrs := make([]string, 0, len(ips)/6)
	for i := 0; i < len(ips); i += 6 {
		ipAddrs = append(ipAddrs, fmt.Sprintf("%d.%d.%d.%d:%d", ips[i], ips[i+1], ips[i+2], ips[i+3], binary.BigEndian.Uint16(ips[i+4:i+6])))
	}
	return ipAddrs
}
