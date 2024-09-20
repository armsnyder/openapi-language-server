package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/armsnyder/openapi-language-server/internal/lsp/jsonrpc"
	"github.com/armsnyder/openapi-language-server/internal/lsp/types"
)

// Server is an LSP server. It handles the I/O and delegates handling of
// requests to a Handler.
type Server struct {
	Reader     io.Reader
	Writer     io.Writer
	Handler    Handler
	ServerInfo types.ServerInfo
}

// Run is a blocking function that reads from the server's Reader, processes
// requests, and writes responses to the server's Writer. It returns an error
// if the server stops unexpectedly.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.Reader)
	scanner.Buffer(nil, 10*1024*1024)
	scanner.Split(jsonrpc.Split)

	log.Println("LSP server started")

	for scanner.Scan() {
		if err := s.handleRequestPayload(scanner.Bytes()); err != nil {
			if errors.Is(err, errShutdown) {
				log.Println("LSP server shutting down")
				return nil
			}

			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequestPayload(payload []byte) (err error) {
	var request types.RequestMessage

	err = json.Unmarshal(payload, &request)
	if err != nil {
		return err
	}

	if request.JSONRPC != "2.0" {
		return errors.New("unknown jsonrpc version")
	}

	if request.Method == "" {
		return errors.New("request is missing a method")
	}

	return s.handleRequest(request)
}

var errShutdown = errors.New("shutdown")

func (s *Server) handleRequest(request types.RequestMessage) error {
	switch request.Method {
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialize
	case "initialize":
		var params types.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid initialize params: %w", err)
		}

		log.Printf("Connected to: %s %s", params.ClientInfo.Name, params.ClientInfo.Version)

		s.write(request, types.InitializeResult{
			Capabilities: s.Handler.Capabilities(),
			ServerInfo:   s.ServerInfo,
		})

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialized
	case "initialized":
		// No-op

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#shutdown
	case "shutdown":
		s.write(request, nil)
		return errShutdown

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_didOpen
	case "textDocument/didOpen":
		var params types.DidOpenTextDocumentParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid textDocument/didOpen params: %w", err)
		}

		if err := s.Handler.HandleOpen(params); err != nil {
			return err
		}

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_didClose
	case "textDocument/didClose":
		var params types.DidCloseTextDocumentParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid textDocument/didClose params: %w", err)
		}

		if err := s.Handler.HandleClose(params); err != nil {
			return err
		}

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_didChange
	case "textDocument/didChange":
		var params types.DidChangeTextDocumentParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid textDocument/didChange params: %w", err)
		}

		if err := s.Handler.HandleChange(params); err != nil {
			return err
		}

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_definition
	case "textDocument/definition":
		var params types.DefinitionParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid textDocument/definition params: %w", err)
		}

		location, err := s.Handler.HandleDefinition(params)
		if err != nil {
			return err
		}

		s.write(request, location)

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_references
	case "textDocument/references":
		var params types.ReferenceParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return fmt.Errorf("invalid textDocument/references params: %w", err)
		}

		locations, err := s.Handler.HandleReferences(params)
		if err != nil {
			return err
		}

		s.write(request, locations)

	default:
		log.Printf("Warning: Request with unknown method %q", request.Method)
	}

	return nil
}

func (s *Server) write(request types.RequestMessage, result any) {
	if err := jsonrpc.Write(s.Writer, types.ResponseMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
