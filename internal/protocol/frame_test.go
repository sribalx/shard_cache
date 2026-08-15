package protocol

import (
	"bytes"
	"testing"
)

// TestFrameRoundTrip verifies that encoding then decoding produces the original frame
// Most important property: encode(decode(frame)) == frame
func TestFrameRoundTrip(t *testing.T) {
	original := Frame{
		Op:    OpSet,
		Key:   []byte("hello"),
		Value: []byte("world"),
	}

	// Encode
	buf := make([]byte, original.EncodedLen())
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != original.EncodedLen() {
		t.Fatalf("Encode returned %d bytes, expected %d", n, original.EncodedLen())
	}

	// Decode header
	op, keyLen, valLen, err := DecodeHeader(buf)
	if err != nil {
		t.Fatalf("DecodeHeader failed: %v", err)
	}

	// Decode payload
	payload := buf[HeaderSize:]
	var decoded Frame
	decoded.Op = op
	decoded.DecodePayload(payload, keyLen, valLen)

	// Verify
	if decoded.Op != original.Op {
		t.Errorf("Op mismatch: got %d, want %d", decoded.Op, original.Op)
	}
	if !bytes.Equal(decoded.Key, original.Key) {
		t.Errorf("Key mismatch: got %q, want %q", decoded.Key, original.Key)
	}
	if !bytes.Equal(decoded.Value, original.Value) {
		t.Errorf("Value mismatch: got %q, want %q", decoded.Value, original.Value)
	}
}

// TestDecodeHeaderInvalidMagic verifies we reject frames with wrong start byte
func TestDecodeHeaderInvalidMagic(t *testing.T) {
	buf := make([]byte, HeaderSize)
	buf[0] = 0x00 // wrong start byte

	_, _, _, err := DecodeHeader(buf)
	if err != ErrInvalidStartByte {
		t.Errorf("expected ErrInvalidStartByte, got %v", err)
	}
}

// TestDecodeHeaderTooShort verifies we reject buffers smaller than header size
func TestDecodeHeaderTooShort(t *testing.T) {
	buf := make([]byte, 4) // too small

	_, _, _, err := DecodeHeader(buf)
	if err != ErrBufferTooSmall {
		t.Errorf("expected ErrBufferTooSmall, got %v", err)
	}
}

// TestEncodeBufferTooSmall verifies Encode rejects undersized buffers
// Caller must provide adequate space, Encode should fail gracefully if not
func TestEncodeBufferTooSmall(t *testing.T) {
	frame := Frame{
		Op:    OpSet,
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	buf := make([]byte, 4) // too small
	_, err := frame.Encode(buf)
	if err != ErrBufferTooSmall {
		t.Errorf("expected ErrBufferTooSmall, got %v", err)
	}
}

// TestEmptyKeyAndValue verifies frames with no key or value work correctly
// Checking because OpOK responses have no key or value, must handle zero-length slices
func TestEmptyKeyAndValue(t *testing.T) {
	original := Frame{
		Op:    OpOK,
		Key:   nil,
		Value: nil,
	}

	// Check encoded length is just the header
	if original.EncodedLen() != HeaderSize {
		t.Fatalf("EncodedLen: got %d, want %d", original.EncodedLen(), HeaderSize)
	}

	// Encode
	buf := make([]byte, original.EncodedLen())
	_, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	op, keyLen, valLen, err := DecodeHeader(buf)
	if err != nil {
		t.Fatalf("DecodeHeader failed: %v", err)
	}

	var decoded Frame
	decoded.Op = op
	decoded.DecodePayload(buf[HeaderSize:], keyLen, valLen)

	// Verify
	if decoded.Op != original.Op {
		t.Errorf("Op mismatch: got %d, want %d", decoded.Op, original.Op)
	}
	if len(decoded.Key) != 0 {
		t.Errorf("Key should be empty, got %q", decoded.Key)
	}
	if len(decoded.Value) != 0 {
		t.Errorf("Value should be empty, got %q", decoded.Value)
	}
}
