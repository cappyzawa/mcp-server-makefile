package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cappyzawa/mcp-server-makefile/internal/mcp"
)

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Set up logging
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	server := mcp.NewServer()
	ctx := context.Background()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Failed to parse request: %v", err)
			continue
		}

		var result interface{}
		var err error

		switch req.Method {
		case "initialize":
			result, err = server.Initialize(ctx, req.Params)
		case "tools/list":
			result, err = server.ListTools(ctx)
		case "tools/call":
			var params struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err = json.Unmarshal(req.Params, &params); err == nil {
				result, err = server.CallTool(ctx, params.Name, params.Arguments)
			}
		default:
			err = fmt.Errorf("unknown method: %s", req.Method)
		}

		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		if err != nil {
			resp.Error = &Error{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}

		respBytes, _ := json.Marshal(resp)
		writer.Write(respBytes)
		writer.WriteByte('\n')
		writer.Flush()
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Fatalf("Error reading input: %v", err)
	}
}