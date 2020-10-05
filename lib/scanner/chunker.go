// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"io"

	"github.com/syncthing/syncthing/lib/protocol"
)

var chunkerFactoryRegistry = make(map[string]ChunkerFactory)

type ChunkerFactory interface {
	// NewChunker returns a new chunker for the data stream in `r`, which is
	// of size `size`. The list of current blocks is optional, if known and
	// passed it may influence the block selection.
	NewChunker(r io.Reader, size int64, curBlocks []protocol.BlockInfo) Chunker
}

type Chunker interface {
	// Chunks returns the number of chunks in the file and true, if known. A
	// chunker that doesn't know the number of chunks until actually reading
	// the stream should return any value and false.
	Chunks() (int, bool)

	// Chunk returns a reader for the next chunk of data, or an error. The
	// error is io.EOF when the last chunk has been processed.
	Chunk() (io.Reader, error)
}

func NewChunkerFactory(name string) ChunkerFactory {
	return chunkerFactoryRegistry[name]
}

func NewDefaultChunkerFactory() ChunkerFactory {
	return chunkerFactoryRegistry["fastcdc"]
}
