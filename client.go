package pieces

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

const BlockSize = 2 << 13

func PeerHandler(peer *Peer, t Tracker) error {
	//pieceLength := t.Torrent.Info.PieceLength
	//blockSize := 256
	//numberOfBlocks := int(math.Ceil(float64(pieceLength) / float64(blockSize)))

	conn, err := NewPeerConn(peer)
	defer conn.Close()
	if err != nil {
		return err
	}
	err = DoHandShake(conn, peer)
	if err != nil {
		return err
	}

	err = ExpressInterest(conn, peer)
	if err != nil {
		return err
	}

	message, err := defaultProtocol.ParseMessage(conn)
	if err != nil {
		return err
	}

	if message.ID == UnChoke {
		slog.Info("sending first block for download...")
		req, err := defaultProtocol.newRequest(0, 0, BlockSize)
		if err != nil {
			return err
		}
		log.Println("request byte slice", "req", req)
		_, err = conn.Write(req)
		if err != nil {
			return err
		}
		message, err = defaultProtocol.ParseMessage(conn)
		if err != nil {
			return err
		}
		slog.Info("message after first download request", "message", message)

		//if message.ID == Have {
		//	pieceIndex := binary.BigEndian.Uint32(message.Payload)
		//	slog.Info("asking for a piece that this peer has", "pieceIndex", pieceIndex)
		//	req, err := defaultProtocol.newRequest(pieceIndex, 0, BlockSize)
		//	if err != nil {
		//		return err
		//	}
		//	_, err = conn.Write(req)
		//	if err != nil {
		//		return err
		//	}
		//	message, err = defaultProtocol.ParseMessage(conn)
		//	if err != nil {
		//		return err
		//	}
		//	slog.Info("message after download request", "pieceIndex", pieceIndex, "message", message)
		//}
	}

	//if message.ID == 1 {
	//	slog.Info("got unchoke message")
	//	var begin = 0
	//	for i := 0; i < numberOfBlocks; i++ {
	//		currBlockSize := min(256, pieceLength-begin)
	//
	//		slog.Info("sending newRequest", "begin", begin, "currBlockSize", currBlockSize)
	//
	//		req, err := defaultProtocol.newRequest(0, uint32(begin), uint32(currBlockSize))
	//		if err != nil {
	//			return err
	//		}
	//		_, err = conn.Write(req)
	//
	//		if err != nil {
	//			return err
	//		}
	//		message, err = defaultProtocol.ParseMessage(conn)
	//		slog.Info("got message", "messageID", messageID, "error", err)
	//		if message.ID == 7 {
	//			begin += blockSize
	//		}
	//		//_ = conn.Close()
	//	}
	//
	//}

	return nil
}

func NewPeerConn(peer *Peer) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", peer.IP, peer.Port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func DoHandShake(conn net.Conn, peer *Peer) (err error) {
	handShake := defaultProtocol.HandShake(peer.InfoHash, peer.PeerID)
	_, err = conn.Write(handShake)
	if err != nil {
		return err
	}
	handShakeResponseBytes := make([]byte, 68)
	if _, err := conn.Read(handShakeResponseBytes); err != nil {
		return err
	}
	handShakeResponse, err := NewHandShake(handShakeResponseBytes)
	if err != nil {
		return err
	}
	slog.Info("handshake completed", "handShakeResponse", handShakeResponse)
	return nil
}

func ExpressInterest(conn net.Conn, _ *Peer) (err error) {
	interest, err := defaultProtocol.interested()
	if err != nil {
		return err
	}
	_, err = conn.Write(interest)
	return nil
}
