// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"bytes"
	"io"

	fastcdc "github.com/jotfs/fastcdc-go"
	"github.com/syncthing/syncthing/lib/protocol"
)

func init() {
	chunkerFactoryRegistry["fastcdc"] = fastcdcChunkerFactory{}
}

type fastcdcChunkerFactory struct{}

func (fastcdcChunkerFactory) NewChunker(r io.Reader, size int64, curBlocks []protocol.BlockInfo) Chunker {
	opts := fastcdc.Options{
		MinSize:     protocol.MinBlockSize,
		AverageSize: protocol.BlockSize(size),
		MaxSize:     protocol.MaxBlockSize,
	}

	// Only returns an error in case of bad opts, which we don't create.
	chunker, err := fastcdc.NewChunker(r, opts)
	if err != nil {
		panic("bug: creating cdc: " + err.Error())
	}

	return &fastcdcChunker{
		chunker: chunker,
	}
}

type fastcdcChunker struct {
	chunker *fastcdc.Chunker
}

func (fastcdcChunker) Chunks() (int, bool) {
	return 0, false
}

func (c fastcdcChunker) Chunk() (io.Reader, error) {
	chn, err := c.chunker.Next()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(chn.Data), nil
}
