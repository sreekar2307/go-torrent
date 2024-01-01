package pieces

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net/http"
	"net/url"
	"time"
)

type Tracker struct {
	Torrent  Torrent
	PeerID   [20]byte
	InfoHash [20]byte
}

func MustCreateTracker(torrent Torrent) Tracker {
	peerID, err := generatePeerID()
	if err != nil {
		panic("failed to generate peerID")
	}
	if err != nil {
		panic("failed to generate info hash")
	}

	return Tracker{
		Torrent:  torrent,
		PeerID:   peerID,
		InfoHash: torrent.Info.Hash,
	}
}

func (t Tracker) Peers(ctx context.Context) ([]*Peer, error) {
	client := http.DefaultClient
	queryParams := make(url.Values)

	queryParams.Set("info_hash", string(t.InfoHash[:]))
	queryParams.Set("peer_id", string(t.PeerID[:]))
	queryParams.Set("port", "6889")
	queryParams.Set("uploaded", "0")
	queryParams.Set("downloaded", "0")
	queryParams.Set("left", fmt.Sprintf("%d", t.Torrent.Info.Length))
	queryParams.Set("compact", "1")
	announceUrl := fmt.Sprintf("%s?%s", t.Torrent.Announce, queryParams.Encode())
	fmt.Println("announceUrl", announceUrl)
	ctx, cancelFunc := context.WithTimeout(ctx, 15*time.Second)
	defer cancelFunc()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, announceUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}
	response, err := bencode.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	trackerResponse := response.(map[string]any)

	if _, ok := trackerResponse["peers"]; !ok {
		panic("peers info not present")
	}

	peersInfoAsString, ok := trackerResponse["peers"].(string)
	if !ok {
		panic("cannot decode peers info only binary model peer decoding is supported")
	}
	peersInfo := []byte(peersInfoAsString)

	if len(peersInfo)%6 != 0 {
		return nil, errors.New("invalid peer info")
	}

	var peers []*Peer
	for i := 0; i < len(peersInfo); i += 6 {
		peer := MustCreatePeerBinaryModel(peersInfo[i : i+6])
		peer.InfoHash = t.InfoHash
		peer.PeerID = t.PeerID
		peers = append(peers, &peer)
	}

	return peers, nil
}

func generatePeerID() ([20]byte, error) {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return [20]byte{}, err
	}
	return peerID, nil
}
