package server

import (
	"io"
	"log"
	"net"

	"shard_cache/internal/protocol"
)

// handleConn reads frames from a TCP connection and processes them

// Each client is a long-lived TCP stream; so this function loops to read frame, process, send response, repeat
// Two-phase read:
//  1. Read exactly 8 bytes (header)
//  2. Parse header to learn payload size
//  3. Read exactly that many bytes using io.ReadFull

// TODO: remove inline processing and submit to worker pool
func (s *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
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
		var response *protocol.Frame
		switch op {
		case protocol.OpGet:
			val, found := s.store.Get(string(frame.Key))
			if found {
				s.metrics.RecordHit()
				response = &protocol.Frame{Op: protocol.OpValue, Value: val}
			} else {
				s.metrics.RecordMiss()
				response = &protocol.Frame{Op: protocol.OpNotFound}
			}
		case protocol.OpSet:
			s.store.Set(string(frame.Key), frame.Value)
			s.metrics.RecordSet()
			response = &protocol.Frame{Op: protocol.OpOK}
		case protocol.OpDelete:
			s.store.Delete(string(frame.Key))
			s.metrics.RecordDelete()
			response = &protocol.Frame{Op: protocol.OpOK}
		default:
			response = &protocol.Frame{Op: protocol.OpError}
		}
		err = sendResponse(conn, response)
		if err != nil {
			return
		}
		s.metrics.RecordBytesOut(int64(len(response.Value)))
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
