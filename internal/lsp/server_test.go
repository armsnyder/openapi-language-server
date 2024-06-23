package lsp_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"go.uber.org/mock/gomock"

	. "github.com/armsnyder/openapiv3-lsp/internal/lsp"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/jsonrpc"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/testutil"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

func TestServer_Basic(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, s *Server, h *testutil.MockHandler)
		requests      []string
		wantResponses []string
	}{
		{
			name: "initialize with default capabilities",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().Capabilities().Return(types.ServerCapabilities{})
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			},
			wantResponses: []string{
				`{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":{"change":0}},"serverInfo":{"name":"test-lsp","version":"0.1.0"}}}`,
			},
		},
		{
			name: "initialize with all capabilities",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().Capabilities().Return(types.ServerCapabilities{
					TextDocumentSync: types.TextDocumentSyncOptions{
						OpenClose: true,
						Change:    types.SyncIncremental,
					},
					DefinitionProvider: true,
					ReferencesProvider: true,
				})
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			},
			wantResponses: []string{
				`{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"textDocumentSync":{"openClose":true,"change":2},"definitionProvider":true,"referencesProvider":true},"serverInfo":{"name":"test-lsp","version":"0.1.0"}}}`,
			},
		},
		{
			name: "initialized",
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"initialized","params":{}}`,
			},
		},
		{
			name: "shutdown",
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"shutdown","params":{}}`,
			},
			wantResponses: []string{
				`{"jsonrpc":"2.0","id":1,"result":null}`,
			},
		},
		{
			name: "textDocument/didOpen",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleOpen(types.DidOpenTextDocumentParams{
					TextDocument: types.TextDocumentItem{
						URI:  "file:///foo.txt",
						Text: "hello world",
					},
				}).Return(nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///foo.txt","text":"hello world"}}}`,
			},
		},
		{
			name: "textDocument/didClose",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleClose(types.DidCloseTextDocumentParams{
					TextDocument: types.TextDocumentIdentifier{URI: "file:///foo.txt"},
				}).Return(nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/didClose","params":{"textDocument":{"uri":"file:///foo.txt"}}}`,
			},
		},
		{
			name: "textDocument/didChange full sync",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleChange(types.DidChangeTextDocumentParams{
					TextDocument: types.TextDocumentIdentifier{URI: "file:///foo.txt"},
					ContentChanges: []types.TextDocumentContentChangeEvent{
						{Text: "hello world"},
					},
				}).Return(nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///foo.txt","version":42},"contentChanges":[{"text":"hello world"}]}}`,
			},
		},
		{
			name: "textDocument/didChange incremental sync",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleChange(types.DidChangeTextDocumentParams{
					TextDocument: types.TextDocumentIdentifier{URI: "file:///foo.txt"},
					ContentChanges: []types.TextDocumentContentChangeEvent{{
						Text:  "carl",
						Range: &types.Range{Start: types.Position{Line: 0, Character: 6}, End: types.Position{Line: 0, Character: 10}},
					}},
				}).Return(nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///foo.txt","version":42},"contentChanges":[{"text":"carl","range":{"start":{"line":0,"character":6},"end":{"line":0,"character":10}}}]}}`,
			},
		},
		{
			name: "textDocument/definition",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleDefinition(types.DefinitionParams{
					TextDocumentPositionParams: types.TextDocumentPositionParams{
						TextDocument: types.TextDocumentIdentifier{URI: "file:///foo.txt"},
						Position:     types.Position{Line: 1, Character: 2},
					},
				}).Return([]types.Location{{
					URI:   "file:///bar.txt",
					Range: types.Range{Start: types.Position{Line: 3, Character: 4}, End: types.Position{Line: 5, Character: 6}},
				}}, nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/definition","params":{"textDocument":{"uri":"file:///foo.txt"},"position":{"line":1,"character":2}}}`,
			},
			wantResponses: []string{
				`{"jsonrpc":"2.0","id":1,"result":[{"uri":"file:///bar.txt","range":{"start":{"line":3,"character":4},"end":{"line":5,"character":6}}}]}`,
			},
		},
		{
			name: "textDocument/references",
			setup: func(t *testing.T, s *Server, h *testutil.MockHandler) {
				h.EXPECT().HandleReferences(types.ReferenceParams{
					TextDocumentPositionParams: types.TextDocumentPositionParams{
						TextDocument: types.TextDocumentIdentifier{URI: "file:///foo.txt"},
						Position:     types.Position{Line: 1, Character: 2},
					},
				}).Return([]types.Location{
					{
						URI:   "file:///bar.txt",
						Range: types.Range{Start: types.Position{Line: 3, Character: 4}, End: types.Position{Line: 5, Character: 6}},
					},
				}, nil)
			},
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///foo.txt"},"position":{"line":1,"character":2}}}`,
			},
			wantResponses: []string{
				`{"jsonrpc":"2.0","id":1,"result":[{"uri":"file:///bar.txt","range":{"start":{"line":3,"character":4},"end":{"line":5,"character":6}}}]}`,
			},
		},
		{
			name: "unknown method",
			requests: []string{
				`{"jsonrpc":"2.0","id":1,"method":"foo","params":{}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := testutil.NewMockHandler(ctrl)
			reader := &bytes.Buffer{}
			writer := &bytes.Buffer{}
			server := Server{
				ServerInfo: types.ServerInfo{
					Name:    "test-lsp",
					Version: "0.1.0",
				},
				Handler: handler,
				Reader:  reader,
				Writer:  writer,
			}

			if tt.setup != nil {
				tt.setup(t, &server, handler)
			}

			send := RPCWriter{Writer: reader}
			for _, req := range tt.requests {
				fmt.Fprint(send, req)
			}

			if err := server.Run(); err != nil {
				t.Fatal("server.Run() error: ", err)
			}

			scanner := bufio.NewScanner(writer)
			scanner.Split(jsonrpc.Split)

			for _, want := range tt.wantResponses {
				if !scanner.Scan() {
					t.Fatal("missing response: ", want)
				}

				if got := scanner.Text(); got != want {
					t.Errorf("got response:\n%s\n\nexpected response:\n%s", got, want)
				}
			}

			if err := scanner.Err(); err != nil {
				t.Fatal("error while reading server responses: ", err)
			}
		})
	}
}

type RPCWriter struct {
	Writer io.Writer
}

func (w RPCWriter) Write(p []byte) (n int, err error) {
	if err := jsonrpc.WritePayload(w.Writer, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

var _ io.Writer = RPCWriter{}
