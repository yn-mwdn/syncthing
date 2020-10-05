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

func init() {
	chunkerFactoryRegistry["standard"] = standardChunkerFactory{}
}

type standardChunkerFactory struct{}

func (s standardChunkerFactory) NewChunker(r io.Reader, size int64, curBlocks []protocol.BlockInfo) Chunker {
	curBlockSize := 0
	if len(curBlocks) > 0 {
		// Size of the first block is either representative of the actual
		// block size, or it's smaller than protocol.MinBlockSize in which
		// case it gets clamped.
		curBlockSize = curBlocks[0].Size
	}
	return NewStandardChunker(r, size, curBlockSize)
}

// NewStandardChunker creates a new fixed-size chunker based on our block
// size calculation, taking into account a slight preference towards
// curBlockSize if it's within reason.
func NewStandardChunker(r io.Reader, size int64, curBlockSize int) Chunker {
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

	// Clamping
	if blockSize < protocol.MinBlockSize {
		blockSize = protocol.MinBlockSize
	} else if blockSize > protocol.MaxBlockSize {
		blockSize = protocol.MaxBlockSize
	}

	return &standardChunker{
		r:         r,
		size:      size,
		pos:       0,
		chunkSize: int64(blockSize),
	}
}

// NewFixedChunker creates a new fixed-sized chunker using exactly the
// specified block size, always.
func NewFixedChunker(r io.Reader, size int64, blockSize int) Chunker {
	return &standardChunker{
		r:         r,
		size:      size,
		pos:       0,
		chunkSize: int64(blockSize),
	}
}

type standardChunker struct {
	r         io.Reader
	size      int64
	pos       int64
	chunkSize int64
}

func (c *standardChunker) Chunks() (int, bool) {
	return int((c.size + 1) / c.chunkSize), true
}

func (c *standardChunker) Chunk() (io.Reader, error) {
	if c.pos >= c.size {
		return nil, io.EOF
	}

	lr := io.LimitReader(c.r, c.chunkSize)
	c.pos += c.chunkSize
	return lr, nil
}
