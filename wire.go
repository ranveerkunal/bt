// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"encoding/binary"
	"fmt"
	"net"
)

var BTHeader [19]byte

func init() {
	copy(BTHeader[:], "BitTorrent protocol")
}

type Header struct {
	Pstrlen  byte
	Pstr     [19]byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerId   [20]byte
}

func (h *Header) String() string {
	return fmt.Sprintf("pstrlen: %d, pstr: %s, reserved: %x, infohash: %x, peerid: %s",
		h.Pstrlen, string(h.Pstr[:]), h.Reserved, h.InfoHash, string(h.PeerId[:]))
}

func Handshake(peer *Peer, ih [20]byte) (chan byte, error) {
	conn, err := net.Dial("tcp", peer.Addr())
	if err != nil {
		return nil, err
	}
	hsReq := Header{
		Pstrlen:  19,
		Pstr:     BTHeader,
		InfoHash: ih,
		PeerId:   BTPeerId,
	}
	err = binary.Write(conn, binary.BigEndian, hsReq)
	if err != nil {
		return nil, err
	}

	btConn := &Conn{peer, ih, conn, make(chan byte)}
	go btConn.Handler()
	return btConn.Stop, nil
}

type Conn struct {
	Peer     *Peer
	InfoHash [20]byte
	Wire     net.Conn
	Stop     chan byte
}

func (c *Conn) Handler() {
	defer func() {
		c.Wire.Close()
		c.Stop <- 'z'
	}()

	hsReq := &Header{}
	err := binary.Read(c.Wire, binary.BigEndian, hsReq)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hsReq)
}
