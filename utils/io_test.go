package utils

import (
	"bytes"
	"io"
	"testing"
)

func TestReadNWriter(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	reader := bytes.NewBufferString("test")
	if ReadNWriter(writer, reader, 4) != nil || writer.String() != "test" {
		t.Errorf("writer: %s", writer.String())
	}

	writer = bytes.NewBuffer(nil)
	reader = bytes.NewBufferString("test")
	if ReadNWriter(writer, reader, 2) != nil || writer.String() != "te" {
		t.Errorf("writer: %s", writer.String())
	} else if ReadNWriter(writer, reader, 2) != nil || writer.String() != "test" {
		t.Errorf("writer: %s", writer.String())
	}

	writer = bytes.NewBuffer(nil)
	reader = bytes.NewBufferString("test")
	if err := ReadNWriter(writer, reader, 5); err == nil {
		t.Error("non-nil")
	} else if err != io.EOF {
		t.Error(err)
	}
}
