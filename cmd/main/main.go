package main

import (
	"context"
	"fmt"
	"pieces"
)

func main() {
	torrent, err := pieces.NewTorrent("samples/lubuntu.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	tracker := pieces.MustCreateTracker(torrent)
	peers, err := tracker.Peers(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	for _, peer := range peers {
		fmt.Printf("peer %s:%s\n", peer.IP, peer.Port)
	}
	if err != nil {
		return
	}
	for _, peer := range peers {
		err = pieces.PeerHandler(peer, tracker)
		fmt.Println(err)
	}

}
