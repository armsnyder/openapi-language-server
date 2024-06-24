package jsonrpc_test

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	. "github.com/armsnyder/openapi-language-server/internal/lsp/jsonrpc"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "single message",
			input: "Content-Length: 17\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			want: []string{
				`{"jsonrpc":"2.0"}`,
			},
		},
		{
			name:  "multiple messages",
			input: "Content-Length: 17\r\n\r\n{\"jsonrpc\":\"2.0\"}Content-Length: 24\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1}",
			want: []string{
				`{"jsonrpc":"2.0"}`,
				`{"jsonrpc":"2.0","id":1}`,
			},
		},
		{
			name:  "extra headers before",
			input: "Extra-Header: foo\r\nContent-Length: 17\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			want: []string{
				`{"jsonrpc":"2.0"}`,
			},
		},
		{
			name:  "extra headers after",
			input: "Content-Length: 17\r\nExtra-Header: foo\r\n\r\n{\"jsonrpc\":\"2.0\"}",
			want: []string{
				`{"jsonrpc":"2.0"}`,
			},
		},
		{
			name:  "incomplete stream",
			input: "Content-Length: 17\r\n\r\n{\"jsonrp",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readers := []io.Reader{
				strings.NewReader(tt.input),
				iotest.OneByteReader(strings.NewReader(tt.input)),
				iotest.HalfReader(strings.NewReader(tt.input)),
			}

			for i, r := range readers {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					scanner := bufio.NewScanner(r)
					scanner.Split(Split)

					var got []string
					for scanner.Scan() {
						got = append(got, scanner.Text())
					}

					if err := scanner.Err(); err != nil {
						t.Fatal("error while scanning input: ", err)
					}

					if len(tt.want) == 0 && len(got) == 0 {
						return
					}

					for i := 0; i < len(tt.want) && i < len(got); i++ {
						if got[i] != tt.want[i] {
							t.Errorf("message #%d: got %q, want %q", i, got[i], tt.want[i])
						}
					}

					if len(got) != len(tt.want) {
						t.Errorf("got %d messages, want %d", len(got), len(tt.want))
					}
				})
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "null",
			input: nil,
			want:  "Content-Length: 4\r\n\r\nnull",
		},
		{
			name:  "empty string",
			input: "",
			want:  "Content-Length: 2\r\n\r\n\"\"",
		},
		{
			name:  "string",
			input: "foo",
			want:  "Content-Length: 5\r\n\r\n\"foo\"",
		},
		{
			name:  "number",
			input: 42,
			want:  "Content-Length: 2\r\n\r\n42",
		},
		{
			name:  "object",
			input: map[string]any{"foo": "bar"},
			want:  "Content-Length: 13\r\n\r\n{\"foo\":\"bar\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if err := Write(buf, tt.input); err != nil {
				t.Fatal(err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
