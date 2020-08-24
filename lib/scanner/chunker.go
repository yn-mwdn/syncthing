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

type Chunker interface {
	Chunks() int
	Chunk() (io.Reader, error)
}

type standardChunker struct {
	r         io.Reader
	size      int64
	pos       int64
	chunkSize int64
}

func newStandardChunker(r io.Reader, size int64, curBlockSize int) *standardChunker {
	blockSize := protocol.BlockSize(size)

	if curBlockSize > 0 {
		// Check if we should retain current block size.
		if blockSize > curBlockSize && blockSize/curBlockSize <= 2 {
			// New block size is larger, but not more than twice larger.
			// Retain.
			blockSize = curBlockSize
		} else if curBlockSize > blockSize && curBlockSize/blockSize <= 2 {
			// Old block size is larger, but not more than twice larger.
			// Retain.
			blockSize = curBlockSize
		}
	}

	return &standardChunker{
		r:         r,
		size:      size,
		pos:       0,
		chunkSize: int64(blockSize),
	}
}

func (c *standardChunker) Chunks() int {
	return int((c.size + 1) / c.chunkSize)
}

func (c *standardChunker) Chunk() (io.Reader, error) {
	if c.pos >= c.size {
		return nil, io.EOF
	}

	lr := io.LimitReader(c.r, c.chunkSize)
	c.pos += c.chunkSize
	return lr, nil
}
