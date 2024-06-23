package analysis

import (
	"bytes"
	"fmt"

	"github.com/armsnyder/openapiv3-lsp/internal/analysis/yaml"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

// Handler implements the LSP handler for the OpenAPI Language Server. It
// contains the business logic for the server.
type Handler struct {
	lsp.NopHandler

	files map[string]*annotatedFile
}

type annotatedFile struct {
	file     lsp.File
	document yaml.Document
}

func (h *Handler) getDocument(uri string) (yaml.Document, error) {
	f := h.files[uri]
	if f == nil {
		return yaml.Document{}, fmt.Errorf("unknown file: %s", uri)
	}

	if f.document.Lines == nil {
		document, err := yaml.Parse(bytes.NewReader(f.file.Bytes()))
		if err != nil {
			return yaml.Document{}, err
		}
		f.document = document
	}

	return f.document, nil
}

func (*Handler) Capabilities() types.ServerCapabilities {
	return types.ServerCapabilities{
		TextDocumentSync: types.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    types.SyncIncremental,
		},
		DefinitionProvider: true,
		ReferencesProvider: true,
	}
}

func (h *Handler) HandleOpen(params types.DidOpenTextDocumentParams) error {
	if h.files == nil {
		h.files = make(map[string]*annotatedFile)
	}

	var f annotatedFile

	f.file.Reset([]byte(params.TextDocument.Text))
	h.files[params.TextDocument.URI] = &f

	return nil
}

func (h *Handler) HandleClose(params types.DidCloseTextDocumentParams) error {
	delete(h.files, params.TextDocument.URI)
	return nil
}

func (h *Handler) HandleChange(params types.DidChangeTextDocumentParams) error {
	f, ok := h.files[params.TextDocument.URI]
	if !ok {
		return fmt.Errorf("unknown file: %s", params.TextDocument.URI)
	}

	for _, change := range params.ContentChanges {
		if err := f.file.ApplyChange(change); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) HandleDefinition(params types.DefinitionParams) ([]types.Location, error) {
	document, err := h.getDocument(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	if params.Position.Line >= len(document.Lines) {
		return nil, nil
	}

	ref := document.Lines[params.Position.Line].Value

	referencedLine := document.Locate(ref)
	if referencedLine == nil {
		return nil, nil
	}

	return []types.Location{{
		URI:   params.TextDocument.URI,
		Range: referencedLine.KeyRange,
	}}, nil
}

func (h *Handler) HandleReferences(params types.ReferenceParams) ([]types.Location, error) {
	document, err := h.getDocument(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	if params.Position.Line >= len(document.Lines) {
		return nil, nil
	}

	ref := document.Lines[params.Position.Line].KeyRef()

	var locations []types.Location

	for _, line := range document.Lines {
		if line.Value == ref {
			locations = append(locations, types.Location{
				URI:   params.TextDocument.URI,
				Range: line.ValueRange,
			})
		}
	}
	return locations, nil
}

var _ lsp.Handler = (*Handler)(nil)
