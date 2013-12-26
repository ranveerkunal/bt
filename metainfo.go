// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ranveerkunal/bencode"

	humanize "github.com/dustin/go-humanize"
)

type file struct {
	Length uint64   `ben:"length"`
	Path   []string `ben:"path"`
	Md5sum string   `ben:"md5sum"`
}

type info struct {
	PieceLength uint64  `ben:"piece length"`
	Pieces      string  `ben:"pieces"`
	Private     uint64  `ben:"private"`
	Name        string  `ben:"name"`
	Length      uint64  `ben:"length"` // Single
	Md5sum      string  `ben:"md5sum"` // Single
	Files       []*file `ben:"files"`  // Multiple
}

type MetaInfo struct {
	Info         *info      `ben:"info"`
	Announce     string     `ben:"announce"`
	AnnounceList [][]string `ben:"announce-list"`
	CreationDate int64      `ben:"creation date"`
	Comment      string     `ben:"comment"`
	CreatedBy    string     `ben:"created by"`
	Encoding     string     `ben:"encoding"`
	infoHash     [20]byte
}

func ReadFile(path string) (*MetaInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rm, err := bencode.Decode(f)
	var info *bencode.RawMessage
	for _, kv := range rm.D {
		if kv.K == "info" {
			info = kv.V
		}
	}
	buf := &bytes.Buffer{}
	info.Marshal(buf)

	mi := &MetaInfo{}
	err = rm.Unmarshal(mi)
	if err != nil {
		return nil, err
	}
	mi.infoHash = sha1.Sum(buf.Bytes())
	return mi, nil
}

func (mi *MetaInfo) InfoHash() [20]byte {
	return mi.infoHash
}

func (mi *MetaInfo) TrackerUrl() (urls []string) {
	trackers := map[string]bool{}
	if len(mi.Announce) > 0 {
		trackers[mi.Announce] = true
	}
	for _, announce := range mi.AnnounceList {
		for _, url := range announce {
			trackers[url] = true
		}
	}

	urls = append(urls, mi.Announce)
	for url, _ := range trackers {
		if url == mi.Announce {
			continue
		}
		urls = append(urls, url)
	}
	return
}

func (mi *MetaInfo) ScraperUrl() (urls []string) {
	trackers := mi.TrackerUrl()
	for _, tracker := range trackers {
		url, err := url.Parse(tracker)
		if err != nil {
			continue
		}
		if strings.HasPrefix(filepath.Base(url.Path), "announce") {
			url.Path = strings.Replace(url.Path, "announce", "scrape", 1)
			urls = append(urls, url.String())
		}
	}
	return
}

func (mi *MetaInfo) String() string {
	s := ""
	if mi.Info != nil {
		s += fmt.Sprintf("info: %x(sha1)\n", mi.infoHash)
		s += fmt.Sprintf("\tpiece length: %d\n", mi.Info.PieceLength)
		s += fmt.Sprintf("\tpieces: suppressed\n")
		s += fmt.Sprintf("\tprivate: %d\n", mi.Info.Private)
		if mi.Info.Name != "" {
			s += fmt.Sprintf("\tname: %q", mi.Info.Name)
		}
		if mi.Info.Length > 0 {
			s += fmt.Sprintf(" (%s)\n", humanize.Bytes(mi.Info.Length))
		} else {
			s += "\n"
		}
		if len(mi.Info.Files) > 0 {
			s += fmt.Sprintf("\tfiles:\n")
		}
		for _, f := range mi.Info.Files {
			s += fmt.Sprintf("\t\tpath: %q (%s)\n", filepath.Join(f.Path...), humanize.Bytes(f.Length))
		}
	}
	if len(mi.Announce) > 0 {
		s += fmt.Sprintf("announce: %q\n", mi.Announce)
	}
	if len(mi.AnnounceList) > 0 {
		s += fmt.Sprintf("announce-list:\n")
	}
	for _, a := range mi.AnnounceList {
		s += fmt.Sprintf("\t%q\n", strings.Join(a, ","))
	}
	if mi.CreationDate > 0 {
		s += fmt.Sprintf("ceation date: %s\n", time.Unix(mi.CreationDate, 0))
	}
	if mi.Comment != "" {
		s += fmt.Sprintf("comment: %q\n", mi.Comment)
	}
	if mi.CreatedBy != "" {
		s += fmt.Sprintf("created by: %q\n", mi.CreatedBy)
	}
	if mi.Encoding != "" {
		s += fmt.Sprintf("encoding: %s\n", mi.Encoding)
	}
	return s
}
