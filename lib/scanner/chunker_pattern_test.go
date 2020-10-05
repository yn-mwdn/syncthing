// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestPatternBreaker(t *testing.T) {
	data := []byte("this is some-*- data with a-*- few-*- break-*- patterns in-*- between")
	b := newPatternBreaker(bytes.NewReader(data), 10, 20, []byte("-*-"))
	for {
		chunk, err := b.Chunk()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		bs, err := ioutil.ReadAll(chunk)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%q\n", bs)
	}
}
