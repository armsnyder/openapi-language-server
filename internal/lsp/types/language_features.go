package types

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#definitionParams.
type DefinitionParams struct {
	TextDocumentPositionParams
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#referenceParams.
type ReferenceParams struct {
	TextDocumentPositionParams
}
