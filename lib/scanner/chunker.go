// Copyright (C) 2020 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package scanner

import (
	"bufio"
	"bytes"
	"io"

	"github.com/syncthing/syncthing/lib/protocol"
)

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

// NewStandardChunkerFactory returns a ChunkerFactory that implements the
// usual Syncthing power-of-two block sizes with slight preference towards
// an existing block size.
func NewStandardChunkerFactory() ChunkerFactory {
	return &standardChunkerFactory{}
}

type standardChunkerFactory struct{}

func (s standardChunkerFactory) NewChunker(r io.Reader, size int64, curBlocks []protocol.BlockInfo) Chunker {
	curBlockSize := 0
	if len(curBlocks) > 0 {
		// Size of the first block is either representative of the actual
		// block size, or it's smaller than protocol.MinBlockSize in which
		// case it gets clamped.
		curBlockSize = int(curBlocks[0].Size)
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

type patternBreaker struct {
	s            *bufio.Scanner
	minChunkSize int
	maxChunkSize int
	pattern      []byte
}

func newPatternBreaker(r io.Reader, min, max int, pattern []byte) *patternBreaker {
	s := bufio.NewScanner(r)
	b := &patternBreaker{
		s:            s,
		minChunkSize: min,
		maxChunkSize: max,
		pattern:      pattern,
	}
	b.s.Split(b.splitFunc)
	b.s.Buffer(make([]byte, min), max)
	return b
}

func (b *patternBreaker) Chunks() (int, bool) {
	return 0, false
}

func (b *patternBreaker) Chunk() (io.Reader, error) {
	if !b.s.Scan() {
		return nil, io.EOF
	}
	return bytes.NewReader(b.s.Bytes()), nil
}

func (b *patternBreaker) splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		if len(data) == 0 {
			return 0, nil, io.EOF
		}
		return len(data), data, nil
	}

	if len(data) < b.minChunkSize {
		return 0, nil, nil
	}

	offset := 0
	for {
		idx := bytes.Index(data[offset:], b.pattern)
		if idx < 0 {
			break
		}
		endPos := offset + idx + len(b.pattern)
		if endPos > b.maxChunkSize {
			return b.maxChunkSize, data[:b.maxChunkSize], nil
		}
		if endPos >= b.minChunkSize {
			return endPos, data[:endPos], nil
		}
		offset += idx + len(b.pattern)
	}

	if len(data) >= b.maxChunkSize {
		return b.maxChunkSize, data[:b.maxChunkSize], nil
	}
	return 0, nil, nil
}
