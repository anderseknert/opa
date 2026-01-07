package topdown

import (
	"bytes"
	"io"
)

var _ io.Writer = (*sinkW)(nil)

// TODO(anderseknert): consider embedding bytes.Buffer directly?
// This interface is getting quite big..

type sinkWriter interface {
	io.Writer
	io.StringWriter
	io.ByteWriter
	String() string
	Bytes() []byte
	Grow(int)
	Len() int
	Reset()
	AvailableBuffer() []byte
}

type sinkW struct {
	buf    *bytes.Buffer
	cancel Cancel
	err    error
}

func newSink(name string, hint int, c Cancel) sinkWriter {
	b := &bytes.Buffer{}
	if hint > 0 {
		b.Grow(hint)
	}

	if c == nil {
		return b
	}

	return &sinkW{
		cancel: c,
		buf:    b,
		err: Halt{
			Err: &Error{
				Code:    CancelErr,
				Message: name + ": timed out before finishing",
			},
		},
	}
}

func (sw *sinkW) Grow(n int) {
	sw.buf.Grow(n)
}

func (sw *sinkW) Write(bs []byte) (int, error) {
	if sw.cancel.Cancelled() {
		return 0, sw.err
	}
	return sw.buf.Write(bs)
}

func (sw *sinkW) WriteByte(b byte) error {
	if sw.cancel.Cancelled() {
		return sw.err
	}
	return sw.buf.WriteByte(b)
}

func (sw *sinkW) WriteString(s string) (int, error) {
	if sw.cancel.Cancelled() {
		return 0, sw.err
	}
	return sw.buf.WriteString(s)
}

func (sw *sinkW) String() string {
	return sw.buf.String()
}

func (sw *sinkW) Bytes() []byte {
	return sw.buf.Bytes()
}

func (sw *sinkW) Len() int {
	return sw.buf.Len()
}

func (sw *sinkW) Reset() {
	sw.buf.Reset()
}

func (sw *sinkW) AvailableBuffer() []byte {
	return sw.buf.AvailableBuffer()
}
