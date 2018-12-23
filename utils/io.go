package utils

import (
	"bytes"
	"io"
)

// ReadN reads the data from io.Reader until n bytes or no incoming data
// if n is equal to or less than 0.
func ReadN(r io.Reader, n int64) (v []byte, err error) {
	buf := bytes.NewBuffer(nil)
	err = ReadNBuffer(buf, r, n)
	return buf.Bytes(), err
}

// ReadNBuffer reads n bytes into buf from r.
func ReadNBuffer(buf *bytes.Buffer, r io.Reader, n int64) error {
	if n < 1 {
		_, err := io.Copy(buf, r)
		return err
	}

	if n < 32768 { // 32KB
		buf.Grow(int(n))
	} else {
		buf.Grow(32768)
	}

	if m, err := io.Copy(buf, io.LimitReader(r, n)); err != nil {
		return err
	} else if m < n {
		return io.EOF
	}
	return nil
}

// ReadNWriter is the same as ReadN, but writes the data to the writer
// from the reader.
func ReadNWriter(w io.Writer, r io.Reader, n int64) (err error) {
	if n > 0 {
		var m int64
		m, err = io.Copy(w, io.LimitReader(r, n))
		if m < n && err == nil {
			err = io.EOF
		}
	} else {
		_, err = io.Copy(w, r)
	}
	return
}
