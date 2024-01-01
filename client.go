package pieces

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net"
	"sync"
)

const BlockSize = 2 << 13

type Client struct {
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
	PieceManager   PieceManager
	Conns          map[[20]byte]net.Conn
	Torrent        *Torrent
}

type pieceWork struct {
	peer   *Peer
	index  int
	hash   [20]byte
	length int
	begin  int
	end    int
}

type pieceResult struct {
	begin int
	end   int
	buf   []byte
}

func NewClient(ctx context.Context, manager PieceManager, tracker *Tracker) (*Client, error) {
	c := &Client{
		AmChoking:      true,
		AmInterested:   false,
		PeerChoking:    true,
		PeerInterested: false,
		PieceManager:   manager,
		Torrent:        tracker.Torrent,
		Conns:          make(map[[20]byte]net.Conn),
	}

	peers, err := tracker.getPeers(ctx)
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		err := c.NewPeerConnection(peer)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) NewPeerConnection(peer *Peer) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", peer.IP, peer.Port))
	if err != nil {
		return err
	}
	err = c.doHandShake(conn, peer)
	if err != nil {
		return err
	}
	err = c.expressInterest(conn, peer)
	if err != nil {
		return err
	}
	message, err := Protocol.ParseMessage(conn)
	if message.ID == UnChokeMessageType {
		c.AmChoking = false
		c.Conns[peer.RemotePeerID] = conn
	}
	return nil
}

func (c *Client) Download(concurrency int) ([]byte, error) {

	ch := make(chan *pieceWork)
	results := make(chan *pieceResult, concurrency)
	buffer := make([]byte, c.Torrent.Info.Length)
	exit := make(chan bool)

	go func() {
		defer close(results)
		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go c.downloadWorker(&wg, ch, results)
		}
		wg.Wait()
	}()

	go func() {
		for result := range results {
			copy(buffer[result.begin:result.end], result.buf)
		}
		exit <- true
	}()

	go func() {
		defer close(ch)
		for {
			piece, peer, err := c.PieceManager.NextPiece()
			if err != nil {
				break
			}
			if piece == nil && peer == nil {
				break
			}
			ch <- &pieceWork{
				peer:   peer,
				index:  piece.Index,
				hash:   piece.Hash,
				length: piece.Length,
				begin:  piece.Begin,
				end:    piece.End,
			}
		}
	}()

	<-exit
	return buffer, nil
}

func (c *Client) downloadWorker(wg *sync.WaitGroup, ch chan *pieceWork, results chan *pieceResult) {
	defer wg.Done()
	for pw := range ch {
		conn, ok := c.Conns[pw.peer.RemotePeerID]
		if !ok {
			log.Println("no connection found for Peer", pw.peer)
			continue
		}
		buf, err := c.downloadPiece(conn, pw.index, pw.length, pw.hash)
		if err != nil {
			log.Println("error downloading piece", err)
			continue
		}
		results <- &pieceResult{
			begin: pw.begin,
			end:   pw.end,
			buf:   buf,
		}
	}
}

func (c *Client) doHandShake(conn net.Conn, peer *Peer) (err error) {
	handShake := Protocol.HandShake(peer.InfoHash, peer.PeerID)
	_, err = conn.Write(handShake)
	if err != nil {
		return err
	}
	handShakeResponseBytes := make([]byte, 68)
	if _, err := conn.Read(handShakeResponseBytes); err != nil {
		return err
	}
	handShakeResponse, err := Protocol.ParseHandShake(handShakeResponseBytes)
	if err != nil {
		return err
	}
	slog.Info("handshake completed", "handShakeResponse", handShakeResponse)
	return nil
}

func (c *Client) expressInterest(conn net.Conn, _ *Peer) (err error) {
	interest, err := Protocol.Interested()
	if err != nil {
		return err
	}
	_, err = conn.Write(interest)
	return nil
}

func (c *Client) downloadPiece(conn net.Conn, pieceIndex int, pieceLength int, _ [20]byte) ([]byte, error) {
	numberOfBlocks := int(math.Ceil(float64(pieceLength) / float64(BlockSize)))
	var begin = 0
	var buffer bytes.Buffer

	for i := 0; i < numberOfBlocks; i++ {
		dataRequested := min(BlockSize, pieceLength-begin)

		slog.Info("sending NewRequest", "pieceIndex", pieceIndex,
			"begin", begin, "dataRequested", dataRequested)

		req, err := Protocol.NewRequest(uint32(pieceIndex), uint32(begin), uint32(dataRequested))
		if err != nil {
			return nil, err
		}
		_, err = conn.Write(req)

		if err != nil {
			return nil, err
		}
		message, err := Protocol.ParseMessage(conn)
		slog.Info("got message", "message", message, "error", err)
		if message.ID == PieceMessageType {
			begin += BlockSize
			buffer.Write(message.Payload)
		} else {
			return nil, fmt.Errorf("unexpected message ID: %d", message.ID)
		}
	}
	return buffer.Bytes(), nil
}
