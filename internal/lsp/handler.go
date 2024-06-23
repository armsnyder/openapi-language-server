package lsp

import "github.com/armsnyder/openapiv3-lsp/internal/lsp/types"

// Handler is an interface for handling LSP requests.
type Handler interface {
	Capabilities() types.ServerCapabilities
	HandleOpen(params types.DidOpenTextDocumentParams) error
	HandleClose(params types.DidCloseTextDocumentParams) error
	HandleChange(params types.DidChangeTextDocumentParams) error
	HandleDefinition(params types.DefinitionParams) ([]types.Location, error)
	HandleReferences(params types.ReferenceParams) ([]types.Location, error)
}

// NopHandler can be embedded in a struct to provide no-op implementations of
// ununsed Handler methods.
type NopHandler struct{}

// Capabilities implements Handler.
func (NopHandler) Capabilities() types.ServerCapabilities {
	return types.ServerCapabilities{}
}

// HandleOpen implements Handler.
func (NopHandler) HandleOpen(types.DidOpenTextDocumentParams) error {
	return nil
}

// HandleClose implements Handler.
func (NopHandler) HandleClose(types.DidCloseTextDocumentParams) error {
	return nil
}

// HandleChange implements Handler.
func (NopHandler) HandleChange(types.DidChangeTextDocumentParams) error {
	return nil
}

// HandleDefinition implements Handler.
func (NopHandler) HandleDefinition(types.DefinitionParams) ([]types.Location, error) {
	return []types.Location{}, nil
}

// HandleReferences implements Handler.
func (NopHandler) HandleReferences(types.ReferenceParams) ([]types.Location, error) {
	return []types.Location{}, nil
}

var _ Handler = NopHandler{}
