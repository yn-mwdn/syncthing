package scanner

import (
	"io"

	"github.com/syncthing/syncthing/lib/protocol"
)

type Chunker interface {
	Chunk() (io.Reader, int, error)
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

func (c *standardChunker) Chunk() (io.Reader, int, error) {
	if c.pos >= c.size {
		return nil, 0, io.EOF
	}

	lr := io.LimitReader(c.r, c.chunkSize)
	c.pos += c.chunkSize
	return lr, int(c.chunkSize), nil
}
