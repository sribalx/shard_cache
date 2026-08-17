package server

import (
	"log"
	"io"
	"net"

	"shard_cache/internal/protocol"
)

// handleConn reads frames from a TCP connection and processes them

// Each client is a long-lived TCP stream; so this function loops to read frame, process, send response, repeat
// Two-phase read:
//  1. Read exactly 8 bytes (header)
//  2. Parse header to learn payload size
//  3. Read exactly that many bytes using io.ReadFull

// TODO: dispatch to the sharded store based on op
// TODO: remove inline processing and submit to worker pool
func (s *Server) handleConn(conn net.Conn) {
	defer func () {_ = conn.Close() }()
	header := make([]byte, protocol.HeaderSize)
	for {
		_, err := io.ReadFull(conn, header)
		if err != nil {
			return
		}
		op, keyLen, valLen, err := protocol.DecodeHeader(header)
		if err != nil {
			log.Printf("failed to decode header: %v", err)
			return
		}
		payloadSize := int(keyLen) + int(valLen)
		var payload []byte
		if payloadSize > 0 {
			payload = make([]byte, payloadSize)
			_,err = io.ReadFull(conn, payload)
			if err != nil {
				return
			}
		}
		frame := &protocol.Frame{Op: op}
		frame.DecodePayload(payload, keyLen, valLen)

		response := &protocol.Frame{Op: protocol.OpOK}
		err = sendResponse(conn, response)
		if err != nil {
			return
		}
	}
}

// sendResponse encodes and writes frame to the connection
// Separate helper because response sending will be in muliple places based on opcodes
func sendResponse(conn net.Conn, response *protocol.Frame) error {
	size := response.EncodedLen()
	buf := make([]byte, size)
	_,err := response.Encode(buf)
	if err != nil {
		log.Printf("failed to encode: %v", err)
		return err
	}
	_,err = conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
