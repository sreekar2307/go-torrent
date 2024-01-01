package pieces

type Piece struct {
	Index  int
	Length int
	Hash   [20]byte
	Begin  int
	End    int
}

type PieceManager interface {
	NextPiece() (*Piece, *Peer, error)
}

type OrderedPieceManager struct {
	CurrIndex int
	Torrent   *Torrent
	Peer      *Peer
}

func (o OrderedPieceManager) NextPiece() (*Piece, *Peer, error) {
	if o.CurrIndex < len(o.Torrent.Info.Pieces) {
		begin := o.CurrIndex * o.Torrent.Info.PieceLength
		end := begin + o.Torrent.Info.PieceLength
		if end > o.Torrent.Info.Length {
			end = o.Torrent.Info.Length
		}
		piece := Piece{
			Index:  o.CurrIndex,
			Length: o.Torrent.Info.PieceLength,
			Hash:   o.Torrent.Info.PieceHashes[o.CurrIndex],
			Begin:  begin,
			End:    end,
		}
		o.CurrIndex++
		return &piece, o.Peer, nil
	}
	return nil, nil, nil
}
