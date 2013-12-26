// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/ranveerkunal/bencode"
)

var BTPeerId [20]byte

func init() {
	ts := strconv.FormatInt(time.Now().UnixNano(), 10)
	copy(BTPeerId[:], fmt.Sprintf("-DD0001-%s", ts))
}

type TrackerReq struct {
	InfoHash   [20]byte `url:"info_hash"`
	PeerId     [20]byte `url:"peer_id"`
	Port       uint16   `url:"port"`
	Uploaded   uint64   `url:"uploaded"`
	Downloaded uint64   `url:"downloaded"`
	Left       uint64   `url:"left"`
	Compact    uint8    `url:"compact"`
}

func (req *TrackerReq) UrlParams() string {
	v := url.Values{}
	typ := reflect.TypeOf(*req)
	val := reflect.Indirect(reflect.ValueOf(req))
	for i := 0; i < typ.NumField(); i++ {
		typf, valf := typ.Field(i), val.Field(i)
		key := typf.Tag.Get("url")
		switch typf.Type.Kind() {
		case reflect.Array:
			v.Add(key, string(valf.Slice(0, valf.Len()).Bytes()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.Add(key, strconv.FormatUint(valf.Uint(), 10))
		}
	}
	return v.Encode()
}

type Peer struct {
	PeerId string `ben:"peer id"`
	IP     string `ben:"ip"`
	Port   uint16 `ben:"port"`
}

func (peer *Peer) Addr() string {
	return fmt.Sprintf("%s:%d", peer.IP, peer.Port)
}

type TrackerRes struct {
	FailureReason  string              `ben:"failure reason"`
	WarningMessage string              `ben:"warning message"`
	Interval       uint64              `ben:"interval"`
	MinInterval    uint64              `ben:"min interval"`
	TrackerId      string              `ben:"tracker id"`
	Complete       uint64              `ben:"complete"`
	Incomplete     uint64              `ben:"incomplete"`
	Peers          *bencode.RawMessage `ben:"peers"`
}

func (res *TrackerRes) ListPeers() (peers []*Peer) {
	if res.Peers.POD != nil {
		l := res.Peers.POD.(string)
		for i := 0; i < len(l); i = i + 6 {
			peer := &Peer{}
			peer.IP = fmt.Sprintf("%d.%d.%d.%d", l[i], l[i+1], l[i+2], l[i+3])
			binary.Read(bytes.NewReader([]byte(l[i+4:i+6])), binary.BigEndian, &peer.Port)
			peers = append(peers, peer)
		}
		return
	}
	res.Peers.Unmarshal(peers)
	return
}

type File struct {
	Complete   uint64 `ben:"complete"`
	Downloaded uint64 `ben:"downloaded"`
	Incomplete uint64 `ben:"incomplete"`
	Name       string `ben:"name"`
}

type Files struct {
	Files map[string]*File `ben:"files"`
}

func (f *Files) String() (s string) {
	for k, v := range f.Files {
		s += fmt.Sprintf("%q: %+v\n", k, *v)
	}
	return
}

func Scrape(path string, ih [][20]byte) (*Files, error) {
	scraper, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	v, err := url.ParseQuery(scraper.RawQuery)
	if err != nil {
		return nil, err
	}

	for _, infoHash := range ih {
		v.Add("info_hash", string(infoHash[:]))
	}
	scraper.RawQuery = v.Encode()
	res, err := http.Get(scraper.String())
	if err != nil {
		return nil, nil
	}
	fmt.Println(scraper.String())

	files := &Files{}
	err = bencode.Unmarshal(res.Body, files)
	if err != nil {
		return nil, err
	}
	return files, nil
}
