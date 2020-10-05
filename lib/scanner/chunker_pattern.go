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

func init() {
	chunkerFactoryRegistry["zeropattern"] = patternBreakerFactory{}
}

type patternBreakerFactory struct{}

func (patternBreakerFactory) NewChunker(r io.Reader, size int64, curBlocks []protocol.BlockInfo) Chunker {
	return newPatternBreaker(r, protocol.MinBlockSize, protocol.MaxBlockSize, []byte{0, 0, 0, 0, 0, 0, 0, 0})
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
