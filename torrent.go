package pieces

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"os"
)

type Info struct {
	Length      int        `bencode:"length"`
	Name        string     `bencode:"name"`
	PieceLength int        `bencode:"piece length"`
	Pieces      string     `bencode:"pieces"`
	PieceHashes [][20]byte `bencode:"-"`
	Hash        [20]byte   `bencode:"-"`
}

func (info *Info) computeHash() ([20]byte, error) {
	var buffer bytes.Buffer
	err := bencode.Marshal(&buffer, *info)
	if err != nil {
		return [20]byte{}, err
	}
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buffer.Bytes()), nil
}

func (info *Info) computePieceHashes() [][20]byte {
	var pieceHashes [][20]byte
	for i := 0; i < len(info.Pieces); i += 20 {
		var hash [20]byte
		copy(hash[:], info.Pieces[i:i+20])
		pieceHashes = append(pieceHashes, hash)
	}
	return pieceHashes
}

type Torrent struct {
	Announce     string `bencode:"announce"`
	Comment      string `bencode:"comment"`
	CreatedBy    string `bencode:"created by"`
	CreationDate int    `bencode:"creation date"`
	Info         Info   `bencode:"info"`
}

func NewTorrent(filePath string) (torrent Torrent, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return torrent, err
	}
	err = bencode.Unmarshal(file, &torrent)
	if err != nil {
		return torrent, nil
	}
	fmt.Println(len(torrent.Info.Pieces))
	if err != nil {
		return torrent, err
	}
	torrent.Info.Hash, err = torrent.Info.computeHash()
	if err != nil {
		return torrent, err
	}
	torrent.Info.PieceHashes = torrent.Info.computePieceHashes()
	return torrent, nil
}
