package protocol

import (
	"testing"
)

// NOTES FROM READING ON FUZZ TESTING
// 
// Go has built-in fuzzing. The fuzzer generates random inputs and feeds them, looking for panics and crashes
// Protocol parser is fuzzed because network input is untrusted, and we can find edge cases that attackers could exploit
// 
// RUN FUZZ TESTS:
//   go test -fuzz=FuzzDecodeHeader ./internal/protocol/   # fuzz one function
//   go test -fuzz=FuzzDecodeHeader -fuzztime=30s ./...    # run for 30 seconds
//
// When bug is found, the crashing input is saved to testdata/fuzz/<FuncName>/<hash> and this becomes a regression test

// FuzzDecodeHeader tests the header parser with random bytes
//
// We're looking for no panics on any input, no out-of-bound slice access, graceful error returns
func FuzzDecodeHeader(f *testing.F) {
	// This seed corpus is the known inputs to start fuzzing from. Fuzzer mutates this to explore input space
	f.Add([]byte{StartByte, OpGet, 0, 3, 0, 0, 0, 0}) // valid header
	f.Add([]byte{})                                    // empty
	f.Add([]byte{StartByte, OpSet})                    // too short
	f.Add([]byte{0xFF, OpGet, 0, 3, 0, 0, 0, 0})       // wrong magic byte

	f.Fuzz(func(t *testing.T, data []byte) {
		op, keyLen, valLen, err := DecodeHeader(data)
		if len(data) < HeaderSize && err == nil {
			t.Error("expected error for short input")
		}
		if len(data) >= HeaderSize && data[0] != StartByte && err == nil {
			t.Error("expected error for invalid magic start byte")
		}
		_, _, _ = op, keyLen, valLen
	})
}

// FuzzDecodePayload tests payload decoding with random data
//
// We're looking predominantly for out-of-bounds access if the keyLen/valLen don't match the actual payload size
func FuzzDecodePayload(f *testing.F) {
	f.Add([]byte("foobar"), uint16(3), uint32(3))    // exact fit: key="foo", val="bar"
	f.Add([]byte("key"), uint16(3), uint32(0))       // key only
	f.Add([]byte("value"), uint16(0), uint32(5))     // value only
	f.Add([]byte{}, uint16(0), uint32(0))            // empty
	f.Add([]byte("short"), uint16(100), uint32(100)) // keyLen+valLen > payload

	f.Fuzz(func(t *testing.T, payload []byte, keyLen uint16, valLen uint32) {
		frame := &Frame{}
		err := frame.DecodePayload(payload, keyLen, valLen)
		required := int(keyLen) + int(valLen)
		if len(payload) < required && err == nil {
			t.Error("expected error for payload smaller than keyLen+valLen")
		}
	})
}

// FuzzRoundTrip tests encode-then-decode with random frames
//
// We're looking for any discrepancies with encoding and decoding with various Frame combinations
func FuzzRoundTrip(f *testing.F) {
	f.Add(byte(OpGet), []byte("mykey"), []byte{})
	f.Add(byte(OpSet), []byte("foo"), []byte("bar"))
	f.Add(byte(OpDelete), []byte("delme"), []byte{})
	f.Add(byte(OpOK), []byte{}, []byte{})
	f.Add(byte(OpValue), []byte{}, []byte("result"))

	f.Fuzz(func(t *testing.T, op byte, key, value []byte) {
		if op > OpNotFound {
			return
		}
		original := &Frame{Op: op, Key: key, Value: value}
		buf := make([]byte, original.EncodedLen())
		original.Encode(buf)

		decodedOp, keyLen, valLen, err := DecodeHeader(buf[:HeaderSize])
		if err != nil {
			t.Fatalf("decode header failed: %v", err)
		}

		decoded := &Frame{Op: decodedOp}
		decoded.DecodePayload(buf[HeaderSize:], keyLen, valLen)

		if decoded.Op != original.Op {
			t.Errorf("op mismatch: got %d, want %d", decoded.Op, original.Op)
		}
		if string(decoded.Key) != string(original.Key) {
			t.Errorf("key mismatch: got %q, want %q", decoded.Key, original.Key)
		}
		if string(decoded.Value) != string(original.Value) {
			t.Errorf("value mismatch: got %q, want %q", decoded.Value, original.Value)
		}
	})
}