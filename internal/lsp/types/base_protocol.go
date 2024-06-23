package types

import (
	"encoding/json"
	"strconv"
)

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#requestMessage.
type RequestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *RequestID      `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#requestMessage.
type RequestID struct {
	IntVal    int
	StringVal string
}

func (r *RequestID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if data[0] == '"' {
		return json.Unmarshal(data, &r.StringVal)
	}

	return json.Unmarshal(data, &r.IntVal)
}

func (r RequestID) MarshalJSON() ([]byte, error) {
	if r.StringVal != "" {
		return json.Marshal(r.StringVal)
	}

	return json.Marshal(r.IntVal)
}

func (r *RequestID) String() string {
	if r == nil {
		return ""
	}

	if r.StringVal != "" {
		return r.StringVal
	}

	return strconv.Itoa(r.IntVal)
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#responseMessage.
type ResponseMessage struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      *RequestID `json:"id"`
	Result  any        `json:"result"`
}
