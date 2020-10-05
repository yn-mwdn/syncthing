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
