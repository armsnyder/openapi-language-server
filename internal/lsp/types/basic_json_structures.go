package types

import "strconv"

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#position.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func (p Position) String() string {
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Character)
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (r Range) String() string {
	return r.Start.String() + "-" + r.End.String()
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocumentItem.
type TextDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocumentIdentifier.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocumentPositionParams.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#location.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}
