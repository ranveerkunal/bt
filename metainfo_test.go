// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bt

import (
	"fmt"
	"testing"
)

func TestSingleFile(t *testing.T) {
	mi, err := ReadFile("./testdata/single_file.torrent")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(mi)
	for _, url := range mi.ScraperUrl() {
		fmt.Printf("scraper: %s\n", url)
	}
}

func TestMultiFile(t *testing.T) {
	mi, err := ReadFile("./testdata/multi_file.torrent")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(mi)
	for _, url := range mi.ScraperUrl() {
		fmt.Printf("scraper: %s\n", url)
	}
}
