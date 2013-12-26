// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"encoding/binary"
	"fmt"
	"io"
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

func Handshake(peer *Peer, ih [20]byte) (chan string, error) {
	conn, err := net.Dial("tcp", peer.Addr())
	if err != nil {
		return nil, err
	}
	hdr := Header{
		Pstrlen:  19,
		Pstr:     BTHeader,
		InfoHash: ih,
		PeerId:   BTPeerId,
	}
	err = binary.Write(conn, binary.BigEndian, hdr)
	if err != nil {
		return nil, err
	}

	btConn := &btConn{peer, ih, conn, make(chan *Message), make(chan string)}
	go btConn.handler()
	return btConn.Ret, nil
}

type btConn struct {
	Peer     *Peer
	InfoHash [20]byte
	Wire     net.Conn
	Msg      chan *Message
	Ret      chan string
}

func readMessages(conn net.Conn, comm chan<- *Message) {
	for {
		msg := &Message{}
		err := binary.Read(conn, binary.BigEndian, &msg.Length)
		if err != nil {
			msg.Error = err
			comm <- msg
			return
		}

		// Keep Alive.
		if msg.Length == 0 {
			comm <- msg
			continue
		}

		// Read Id.
		binary.Read(conn, binary.BigEndian, &msg.Id)
		if err != nil {
			msg.Error = err
			comm <- msg
			return
		}
		if msg.Length == 1 {
			continue // No Payload.
		}

		// Read Payload.
		msg.Payload = make([]byte, msg.Length-1, msg.Length-1)
		_, err = io.ReadFull(conn, msg.Payload)
		if err != nil {
			msg.Error = err
			comm <- msg
			return
		}
	}
}

func (c *btConn) handler() {
	defer func() {
		c.Wire.Close()
	}()

	hdr := &Header{}
	err := binary.Read(c.Wire, binary.BigEndian, hdr)
	if err != nil {
		c.Ret <- err.Error()
		return
	}

	go readMessages(c.Wire, c.Msg)
	for {
		select {
		case msg := <-c.Msg:
			fmt.Println(*msg)
		case ret := <-c.Ret:
			if ret == "CLOSE" {
				c.Ret <- "CLOSED"
				return
			}
		}
	}
	c.Ret <- "OK"
}

type Message struct {
	Length  uint32
	Id      byte
	Payload []byte
	Error   error
}

var (
	KeepAlive     = &Message{Length: 0}
	Choke         = &Message{Length: 1, Id: 0}
	Unchoke       = &Message{Length: 1, Id: 1}
	Interested    = &Message{Length: 1, Id: 2}
	NotInterested = &Message{Length: 1, Id: 3}
	Have          = &Message{Length: 5, Id: 4}
	BitField      = &Message{Length: 1, Id: 5}
	Request       = &Message{Length: 13, Id: 6}
	Piece         = &Message{Length: 9, Id: 7}
	Cancel        = &Message{Length: 13, Id: 8}
	Port          = &Message{Length: 3, Id: 9}
)
