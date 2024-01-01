package main

import (
	"context"
	"fmt"
	"os"
	"pieces"
)

func main() {
	torrent, err := pieces.NewTorrent("samples/lubuntu.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	tracker, err := pieces.NewTracker(&torrent)
	if err != nil {
		return
	}

	pieceManager := pieces.OrderedPieceManager{
		CurrIndex: 0,
		Torrent:   &torrent,
		Peer:      tracker.Peers[0],
	}

	client, err := pieces.NewClient(context.Background(), pieceManager, tracker)

	download, err := client.Download(10)
	if err != nil {
		return
	}
	_ = os.WriteFile("lubuntu.iso", download, 0644)

}
