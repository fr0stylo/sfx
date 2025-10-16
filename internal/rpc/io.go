package rpc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// WriteDelimited writes a length-delimited protobuf message to w.
func WriteDelimited(w io.Writer, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(payload)))

	if _, err := w.Write(lenBuf[:n]); err != nil {
		return fmt.Errorf("write length: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}

	return nil
}

// ReadDelimited reads a length-delimited protobuf message from r into msg.
func ReadDelimited(r io.Reader, msg proto.Message) error {
	br := bufio.NewReader(r)

	length, err := binary.ReadUvarint(br)
	if err != nil {
		return fmt.Errorf("read length: %w", err)
	}

	payload := make([]byte, int(length))
	if _, err := io.ReadFull(br, payload); err != nil {
		return fmt.Errorf("read payload: %w", err)
	}

	if err := proto.Unmarshal(payload, msg); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return nil
}
