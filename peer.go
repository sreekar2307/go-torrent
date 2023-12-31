package pieces

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Peer struct {
	IP           string
	Port         string
	PeerID       [20]byte
	InfoHash     [20]byte
	RemotePeerID [20]byte
}

func MustCreatePeerBinaryModel(info []byte) Peer {
	if len(info) < 6 {
		panic("decoding ip address and port for Peer failed")
	}
	return Peer{
		IP:   net.IPv4(info[0], info[1], info[2], info[3]).String(),
		Port: fmt.Sprintf("%d", binary.BigEndian.Uint16(info[4:6])),
	}
}

func padZeros(n int) []byte {
	var res []byte
	for i := 0; i < n; i++ {
		res = append(res, byte(0))
	}
	return res
}
