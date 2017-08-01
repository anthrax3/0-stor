package compress

import (
	"bytes"
	"io"

	"github.com/pierrec/lz4"

	"github.com/zero-os/0-stor/client/lib/block"
)

type lz4Writer struct {
	w block.Writer
}

func newLz4Writer(w block.Writer) *lz4Writer {
	return &lz4Writer{
		w: w,
	}
}

func (lw lz4Writer) WriteBlock(data []byte) block.WriteResponse {
	buf := bytes.NewBuffer(nil)

	rd := bytes.NewReader(data)

	comp := lz4.NewWriter(buf)
	comp.Header.BlockDependency = true

	// compress the data
	n, err := io.Copy(comp, rd)
	if err != nil {
		return block.WriteResponse{
			Written: int(n),
			Err:     err,
		}
	}

	// flush and close the compressor
	if err := comp.Close(); err != nil {
		return block.WriteResponse{
			Written: int(n),
			Err:     err,
		}
	}

	// return the valuo of our output buffer and a possible error
	return lw.w.WriteBlock(buf.Bytes())
}

// lz4Reader wraps lz4.Reader to conform to Decompressor interface
type lz4Reader struct {
}

func newLz4Reader() *lz4Reader {
	return &lz4Reader{}
}

func (lr *lz4Reader) ReadBlock(data []byte) ([]byte, error) {
	br := bytes.NewReader(data)
	rd := lz4.NewReader(br)

	buf := new(bytes.Buffer)

	_, err := io.Copy(buf, rd)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}