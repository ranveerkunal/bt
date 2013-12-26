// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ranveerkunal/bencode"
)

func TestTracker(t *testing.T) {
	mi, err := ReadFile("./testdata/single_file.torrent")
	if err != nil {
		t.Fatal(err)
	}

	treq := &TrackerReq{
		InfoHash:   mi.InfoHash(),
		PeerId:     BTPeerId(),
		Port:       6885,
		Uploaded:   0,
		Downloaded: 0,
		Left:       0,
		Compact:    1,
	}

	url := fmt.Sprintf("%s?%s", mi.Announce, treq.UrlParams())
	fmt.Println(url)

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()
	tres := &TrackerRes{}
	err = bencode.Unmarshal(res.Body, tres)
	if err != nil {
		t.Fatal(err)
	}
	for _, peer := range tres.ListPeers() {
		fmt.Printf("%s:%s:%d\n", peer.PeerId, peer.IP, peer.Port)
	}
	files, err := Scrape(mi.ScraperUrl()[0], [][20]byte{mi.InfoHash()})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(files)
}
