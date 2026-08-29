package protocol

import (
	"encoding/binary"
	"errors"
)

// Common errors returned by frame operations
// Callers can check error identity with errors.Is(), avoids string matching ("if err.Error() == ...")
var (
	ErrInvalidStartByte = errors.New("invalid start byte")
	ErrBufferTooSmall   = errors.New("buffer too small for frame")
	ErrPayloadTooSmall  = errors.New("payload smaller than keyLen + valLen")
)

// Frame represents a single protocol message
type Frame struct {
	Op    byte   // operation code
	Key   []byte // key bytes (empty for responses like OpOK)
	Value []byte // value bytes (empty for GET requests, OpOK responses)
}

// EncodedLen returns the total number of bytes needed to encode this frame
// Allows caller to control memory, determining how big a buffer to provide to support memory pooling
func (f *Frame) EncodedLen() int {
	return HeaderSize + len(f.Key) + len(f.Value)
}

// Encode writes the frame into buf and returns the number of bytes written
// buf is provided here to support future zero allocation pattern
//
// RETURNS:
// - No. of bytes written
// - Error if buf is too small
func (f *Frame) Encode(buf []byte) (int, error) {
	// Check if buf is large enough
	if len(buf) < f.EncodedLen() {
		return 0, ErrBufferTooSmall
	}
	buf[0] = StartByte
	buf[1] = f.Op
	binary.BigEndian.PutUint16(buf[2:4], uint16(len(f.Key)))
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(f.Value)))
	copy(buf[8:], f.Key)
	copy(buf[8+len(f.Key):], f.Value)
	return f.EncodedLen(), nil
}

// DecodeHeader parses the 8-byte header from buf
//
// RETURNS:
// - op: the operation code
// - keyLen: no. of bytes in the key
// - valLen: no. of bytes in the value
// - err: ErrInvalidStartByte if start byte wrong, ErrBufferTooSmall if < 8 bytes
func DecodeHeader(buf []byte) (op byte, keyLen uint16, valLen uint32, err error) {
	// Check if buf length is at least as long as HeaderSize
	if len(buf) < HeaderSize {
		return 0, 0, 0, ErrBufferTooSmall
	}
	if buf[0] != StartByte {
		return 0, 0, 0, ErrInvalidStartByte
	}
	op = buf[1]
	keyLen = binary.BigEndian.Uint16(buf[2:4])
	valLen = binary.BigEndian.Uint32(buf[4:8])
	return op, keyLen, valLen, nil
}

// DecodePayload extracts Key and Value from payload buffer into the Frame
// Note: zero-copy, slices point into payload. Be careful when reusing payload buffer.
func (f *Frame) DecodePayload(payload []byte, keyLen uint16, valLen uint32) error {
  	required := int(keyLen) + int(valLen)
	if len(payload) < required {
		return ErrPayloadTooSmall
	}
	f.Key = payload[0:keyLen]
	f.Value = payload[keyLen : uint32(keyLen)+valLen]
	return nil
}
