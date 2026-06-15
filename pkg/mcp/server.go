package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      interface{}     `json:"id"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	Id      interface{} `json:"id,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Server struct {
	handlers map[string]func(json.RawMessage) (interface{}, error)
	schemas  map[string]map[string]interface{}
	mode     string
}

func NewServer(mode string) *Server {
	return &Server{
		handlers: make(map[string]func(json.RawMessage) (interface{}, error)),
		schemas:  make(map[string]map[string]interface{}),
		mode:     mode,
	}
}

func (s *Server) RegisterHandler(name string, schema map[string]interface{}, fn func(json.RawMessage) (interface{}, error)) {
	s.handlers[name] = fn
	s.schemas[name] = schema
}

func (s *Server) Start() error {
	switch s.mode {
	case "stdio":
		return s.runStdio()
	case "http":
		return s.runHTTP()
	default:
		return fmt.Errorf("unsupported mode: %s", s.mode)
	}
}

func (s *Server) runStdio() error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req Request

		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Fprintf(os.Stderr, "decode error: %v\n", err)
			continue
		}

		resp, ok := s.handleRequest(req)
		if !ok {
			continue
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
		}
	}
}

func (s *Server) runHTTP() error {
	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		var req Request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, ok := s.handleRequest(req)

		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	fmt.Fprintln(os.Stderr, "MCP HTTP server running on :3001")
	return http.ListenAndServe(":3001", nil)
}

func (s *Server) handleRequest(req Request) (*Response, bool) {
	if req.Id == nil {
		return nil, false
	}

	resp := &Response{
		Jsonrpc: "2.0",
		Id:      req.Id,
	}

	switch req.Method {

	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
				"resources": map[string]interface{}{},
				"prompts":   map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "jira-mcp",
				"version": "1.0.0",
			},
		}

	case "tools/list", "list_tools":
		tools := make([]map[string]interface{}, 0)

		for name := range s.handlers {
			schema := s.schemas[name]
			if schema == nil {
				schema = map[string]interface{}{"type": "object"}
			}
			tools = append(tools, map[string]interface{}{
				"name":        name,
				"description": fmt.Sprintf("Tool %s", name),
				"inputSchema": schema,
			})
		}

		resp.Result = map[string]interface{}{
			"tools": tools,
		}

	case "tools/call", "call_tool":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{-32602, "invalid params"}
			return resp, true
		}

		handler, ok := s.handlers[params.Name]
		if !ok {
			resp.Error = &RPCError{-32601, "unknown tool"}
			return resp, true
		}

		result, err := handler(params.Arguments)
		if err != nil {
			resp.Error = &RPCError{-32000, err.Error()}
			return resp, true
		}

		resultJSON, _ := json.Marshal(result)
		resp.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
			"isError": false,
		}

	default:
		resp.Error = &RPCError{-32601, "method not found"}
	}

	return resp, true
}

func (s *Server) Stop() {}
