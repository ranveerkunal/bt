// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"encoding/binary"
	"net"
)

type Conn struct {
	Peer     *Peer
	InfoHash [20]byte
	Wire     net.Conn
	Stop     chan byte
}

type handshakeReq struct {
	pstrlen  byte
	pstr     [19]byte
	reserved [8]byte
	infoHash [20]byte
	peerId   [20]byte
}

func Handshake(peer *Peer, ih [20]byte, handler func(*Conn)) (chan byte, error) {
	conn, err := net.Dial("tcp", peer.Addr())
	if err != nil {
		return nil, err
	}
	hsReq := handshakeReq{
		pstrlen:  19,
		infoHash: ih,
		peerId:   BTPeerId,
	}
	copy(hsReq.pstr[:], "BitTorrent protocol")
	err = binary.Write(conn, binary.BigEndian, hsReq)
	if err != nil {
		return nil, err
	}

	btConn := &Conn{peer, ih, conn, make(chan byte)}
	go handler(btConn)
	return btConn.Stop, nil
}
