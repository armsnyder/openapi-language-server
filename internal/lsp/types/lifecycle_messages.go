package types

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initializeParams.
type InitializeParams struct {
	ClientInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initializeResult.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#serverCapabilities.
type ServerCapabilities struct {
	TextDocumentSync   TextDocumentSyncOptions `json:"textDocumentSync"`
	DefinitionProvider bool                    `json:"definitionProvider,omitempty"`
	ReferencesProvider bool                    `json:"referencesProvider,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initializeResult.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
