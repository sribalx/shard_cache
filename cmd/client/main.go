// Package main is a simple CLI client to manually test that the cache server actually works 
//
// Usage: client <get|set|del> <key> [value]
package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"shard_cache/internal/protocol"
)

const serverAddr = "localhost:9000"

// main parses CLI args and sends a request to the server
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client <get|set|del> <key> [value]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	key := os.Args[2]

	var frame *protocol.Frame

	switch cmd {
	case "get":
		frame = &protocol.Frame{Op: protocol.OpGet, Key: []byte(key)}
	case "set":
		if len(os.Args) < 4 {
			fmt.Println("Set requires a value. Usage: client set <key> <value>")
			os.Exit(1)
		}
		frame = &protocol.Frame{Op: protocol.OpSet, Key: []byte(key), Value: []byte(os.Args[3])}
	case "del":
		frame = &protocol.Frame{Op: protocol.OpDelete, Key: []byte(key)}
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		os.Exit(1)
	}

	resp, err := sendRequest(frame)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	printResponse(resp)
}

// sendRequest connects to the server, sends a frame, and reads the response
func sendRequest(frame *protocol.Frame) (*protocol.Frame, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()


	buf := make([]byte, frame.EncodedLen())
	_, err = frame.Encode(buf)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(buf)
	if err != nil {
		return nil, err
	}
	
	
	header := make([]byte, protocol.HeaderSize)
	_, err = io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	op, keyLen, valLen, err := protocol.DecodeHeader(header)
	if err != nil {
		return nil, err
	}
	
	payloadSize := int(keyLen) + int(valLen)
	var payload []byte
	if payloadSize > 0 {
		payload = make([]byte, payloadSize)
		_,err = io.ReadFull(conn, payload)
		if err != nil {
			return nil, err
		}
	}

	response := &protocol.Frame{Op:op}
	err = response.DecodePayload(payload, keyLen, valLen)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// printResponse displays the server's response in human-readable form
func printResponse(resp *protocol.Frame) {
	op := resp.Op

	switch op {
	case protocol.OpOK: 
		fmt.Println("OK")
	case protocol.OpValue:
		fmt.Printf("Value: %s\n", string(resp.Value))
	case protocol.OpNotFound:
		fmt.Println("NOT FOUND")
	case protocol.OpError:
		fmt.Println("ERROR")
	}
}
