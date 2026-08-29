package server

import (
	"io"
	"log"
	"net"

	"shard_cache/internal/protocol"
	"shard_cache/internal/pool"
)

// handleConn reads frames from a TCP connection and processes them

// Each client is a long-lived TCP stream; so this function loops to read frame, process, send response, repeat
// Two-phase read:
//  1. Read exactly 8 bytes (header)
//  2. Parse header to learn payload size
//  3. Read exactly that many bytes using io.ReadFull
// 
// Note: we only pool the header, because if we pool the buffer and return it before the worker
// processes, the frame data becomes garbage
func (s *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)
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
			_, err = io.ReadFull(conn, payload)
			if err != nil {
				return
			}
		}
		s.metrics.RecordBytesIn(int64(payloadSize))
		frame := &protocol.Frame{Op: op}
		frame.DecodePayload(payload, keyLen, valLen)
		if !s.workers.Submit(Job{Conn: conn, Frame: frame}) {
			continue
		}
		
	}
}

// sendResponse encodes and writes frame to the connection
// Separate helper because response sending will be in muliple places based on opcodes
func sendResponse(conn net.Conn, response *protocol.Frame) error {
	size := response.EncodedLen()
	buf := make([]byte, size)
	_, err := response.Encode(buf)
	if err != nil {
		log.Printf("failed to encode: %v", err)
		return err
	}
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
