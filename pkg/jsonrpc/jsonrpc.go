package jsonrpc

import (
	"encoding/json"
)

const (
	ErrorMethodNotFound = -32601
	ErrorInvalidParams  = -32602
	ErrorInternal       = -32603
)

type ResponseError struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   Error           `json:"error"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponsePayload struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
}

type RequestPayload struct {
	Version string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

func NewJSONRPCResponseError(id json.RawMessage, errCode int, msg string) ResponseError {
	return ResponseError{
		Version: "2.0",
		ID:      id,
		Error: Error{
			Code:    errCode,
			Message: msg,
		},
	}
}
