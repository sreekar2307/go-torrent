package pieces

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
)

type protocol struct{}

type HandShake struct {
	PstrLen  byte
	Pstr     [19]byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

const KeepAlive byte = 10

const (
	Choke = byte(iota)
	UnChoke
	Interested
	NotInterested
	Have
	BitField
	Request
	Piece
	Cancel
	Port
)

type Message struct {
	Len     uint32
	ID      byte
	Payload []byte
}

func NewMessage(id byte, payload []byte) (m Message, err error) {
	if id > KeepAlive || id < Choke {
		err = errors.New("invalid message id")
		return
	}
	if id == KeepAlive {
		m.Len = 0
		m.Payload = nil
		return
	}
	m.ID = id
	m.Payload = payload
	m.Len = uint32(len(payload) + 1)
	return
}

func (m Message) Serialize() []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, m.Len)
	if m.Len > 0 {
		res = append(res, m.ID)
	}
	if m.Payload != nil && len(m.Payload) > 0 {
		res = append(res, m.Payload...)
	}
	return res
}

func (p protocol) HandShake(infoHash [20]byte, peerID [20]byte) []byte {
	res := []byte{byte(19)}
	res = append(res, []byte("BitTorrent protocol")...)
	res = append(res, padZeros(8)...)
	res = append(res, infoHash[:]...)
	res = append(res, peerID[:]...)
	return res
}

func NewHandShake(val []byte) (handShake HandShake, err error) {
	err = binary.Read(bytes.NewReader(val), binary.BigEndian, &handShake)
	return
}

//
//func (p protocol) keepAlive() []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 0)
//	return res
//}
//
//func (p protocol) choke() []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 1)
//	res = append(res, byte(0))
//	return res
//}
//
//func NewChoke(val []byte) (choke Choke, err error) {
//	err = binary.Read(bytes.NewReader(val), binary.BigEndian, &choke)
//	return
//}
//
//func (p protocol) unchoke() []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 1)
//	res = append(res, byte(1))
//	return res
//}
//

func (p protocol) interested() ([]byte, error) {
	m, err := NewMessage(Interested, nil)
	if err != nil {
		return nil, err
	}
	return m.Serialize(), nil
}

//
//func NewIntrested(val []byte) (intersted Intrested, err error) {
//	err = binary.Read(bytes.NewReader(val), binary.BigEndian, &intersted)
//	return
//}
//
//func (p protocol) notIntrested() []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 1)
//	res = append(res, byte(3))
//	return res
//}
//
//func (p protocol) have(index uint32) []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 5)
//	res = append(res, byte(4))
//	binary.BigEndian.PutUint32(res, index)
//	return res
//}
//
//func (p protocol) bitField(bitArr []byte) []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 1+uint32(len(bitArr)))
//	res = append(res, byte(5))
//	res = append(res, bitArr...)
//	return res
//}
//

func (p protocol) newRequest(index, begin, length uint32) ([]byte, error) {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)
	m, err := NewMessage(Request, payload)
	if err != nil {
		return nil, err
	}
	slog.Info("new request", "message", m)
	return m.Serialize(), nil
}

//
//func (p protocol) piece(index, begin uint32, block []byte) []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 9+uint32(len(block)))
//	res = append(res, byte(7))
//	binary.BigEndian.PutUint32(res, index)
//	binary.BigEndian.PutUint32(res, begin)
//	res = append(res, block...)
//	return res
//}
//
//
//func (p protocol) cancel(index, begin, length uint32) []byte {
//	res := make([]byte, 4)
//	binary.BigEndian.PutUint32(res, 13)
//	res = append(res, byte(8))
//	binary.BigEndian.PutUint32(res, index)
//	binary.BigEndian.PutUint32(res, begin)
//	binary.BigEndian.PutUint32(res, length)
//	return res
//}

var defaultProtocol = protocol{}

func (p protocol) ParseMessage(reader io.Reader) (message Message, err error) {
	messageLenAsBytes := make([]byte, 4)
	_, err = reader.Read(messageLenAsBytes)
	if err != nil {
		return
	}
	message.Len = binary.BigEndian.Uint32(messageLenAsBytes)

	if message.Len > 0 {
		messageIDAsBytes := make([]byte, 1)
		_, err = reader.Read(messageIDAsBytes)
		if err != nil {
			return
		}
		message.ID = messageIDAsBytes[0]
	}

	if message.Len > 1 {
		message.Payload = make([]byte, message.Len-1)
		_, err = reader.Read(message.Payload)
		if err != nil {
			return
		}
	}

	slog.Info("read message", "message", message)

	return
}
