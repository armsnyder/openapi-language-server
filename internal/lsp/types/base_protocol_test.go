package types_test

import (
	"encoding/json"
	"testing"

	. "github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

func TestResponseMessage_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		message ResponseMessage
		want    string
	}{
		{
			name:    "empty",
			message: ResponseMessage{JSONRPC: "2.0"},
			want:    `{"jsonrpc":"2.0","id":null,"result":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatal(err)
			}
			if string(s) != tt.want {
				t.Errorf("got %s, want %s", s, tt.want)
			}
		})
	}
}
