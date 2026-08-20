// Package protocol defines the binary wire format for client-server communication.
// Custom binary protocol with fixed byte positions, no parsing, minimal overhead.
// Ideal for cache (100k requests per second).
package protocol

// StartByte is the first byte for every valid frame.
// 0xCA is arbitrary.
const StartByte byte = 0xCA

// Request opcodes are what the client wants to do.
// Single byte suffices for our use case.
const (
	OpGet    byte = iota + 1 // 0x01 retrieve value for key
	OpSet                    // 0x02 store k-v pair
	OpDelete                 // 0x03 remove key
)

// Response opcodes are what the server will send back.
// start at 0x80 to have a clear split between request and response.
const (
	OpOK       byte = 0x80 // operation success, no return value
	OpError    byte = 0x81 // operation failed, returning error code
	OpValue    byte = 0x82 // returning value as a response to a get call
	OpNotFound byte = 0x83 // cache miss
)

// HeaderSize is fixed byte size of every frame's header.
// 1 byte for StartByte, 1 byte for OpCode, 2 bytes for KeyLen, 4 bytes for ValLen
// KeyLen is big-endian uint16, max 64KB
// ValLen is big-endian uint32, max 4GB
const HeaderSize = 8
