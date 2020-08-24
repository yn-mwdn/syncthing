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
	return b
}

func (b *patternBreaker) Chunks() int {
	return -1
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

	if len(data) > b.maxChunkSize {
		return b.maxChunkSize, data[:b.maxChunkSize], nil
	}
	return 0, nil, nil
}
