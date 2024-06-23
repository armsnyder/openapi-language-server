package yaml

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

// Document represents a YAML document.
type Document struct {
	Lines []*Line
	Root  map[string]*Line
}

// Locate finds a line in the document by its JSON reference URI.
func (s Document) Locate(ref string) *Line {
	split := strings.Split(ref, "/")
	if len(split) < 2 {
		return nil
	}

	cur := s.Root[split[1]]
	if cur == nil {
		return nil
	}

	for _, key := range split[2:] {
		cur = cur.Children[key]
		if cur == nil {
			return nil
		}
	}

	return cur
}

// Line represents a line in a YAML document.
type Line struct {
	Parent     *Line
	Children   map[string]*Line
	Key        string
	Value      string
	KeyRange   types.Range
	ValueRange types.Range
}

// KeyRef returns the JSON reference URI that describes the key on this line.
func (e *Line) KeyRef() string {
	keys := []string{}

	for cur := e; cur != nil; cur = cur.Parent {
		keys = append(keys, cur.Key)
	}

	var b strings.Builder

	b.WriteString("#/")

	for i := len(keys) - 1; i >= 0; i-- {
		b.WriteString(keys[i])

		if i > 0 {
			b.WriteByte('/')
		}
	}

	return b.String()
}

type lineWithIndent struct {
	line   *Line
	indent int
}

// Parse parses a YAML document from a reader, using best-effort. The YAML does
// not need to be syntactically valid.
func Parse(r io.Reader) (Document, error) {
	parentStack := []lineWithIndent{}
	scanner := bufio.NewScanner(r)
	document := Document{
		Root: map[string]*Line{},
	}

	for lineNum := 0; scanner.Scan(); lineNum++ {
		line := parseLine(scanner.Bytes(), lineNum)
		document.Lines = append(document.Lines, line.line)

		for len(parentStack) > 0 && parentStack[len(parentStack)-1].indent >= line.indent {
			parentStack = parentStack[:len(parentStack)-1]
		}

		if len(parentStack) == 0 {
			document.Root[line.line.Key] = line.line
			parentStack = append(parentStack, line)
			continue
		}

		parent := parentStack[len(parentStack)-1]
		line.line.Parent = parent.line
		if parent.line.Children == nil {
			parent.line.Children = map[string]*Line{}
		}
		parent.line.Children[line.line.Key] = line.line
		parentStack = append(parentStack, line)
	}

	if err := scanner.Err(); err != nil {
		return Document{}, err
	}

	return document, nil
}

func parseLine(s []byte, lineNum int) lineWithIndent {
	result := lineWithIndent{
		line: &Line{},
	}

	result.indent = bytes.IndexFunc(s, func(ch rune) bool {
		return ch != ' '
	})
	if result.indent == -1 {
		result.indent = len(s)
	}

	keyEnd := bytes.Index(s, []byte(":"))
	if keyEnd == -1 {
		return result
	}

	result.line.Key = string(s[result.indent:keyEnd])
	result.line.KeyRange = types.Range{
		Start: types.Position{Line: lineNum, Character: result.indent},
		End:   types.Position{Line: lineNum, Character: keyEnd},
	}

	valueStart := bytes.IndexFunc(s[keyEnd+1:], func(ch rune) bool {
		return ch != ' '
	})
	if valueStart == -1 {
		return result
	}
	valueStart += keyEnd + 1
	if valueStart >= len(s) {
		return result
	}

	if s[valueStart] == '"' || s[valueStart] == '\'' {
		valueEnd := bytes.LastIndex(s, s[valueStart:valueStart+1])
		if valueEnd <= valueStart {
			return result
		}

		result.line.Value = string(s[valueStart+1 : valueEnd])
		result.line.ValueRange = types.Range{
			Start: types.Position{Line: lineNum, Character: valueStart + 1},
			End:   types.Position{Line: lineNum, Character: valueEnd},
		}

		return result
	}

	result.line.Value = string(s[valueStart:])
	result.line.ValueRange = types.Range{
		Start: types.Position{Line: lineNum, Character: valueStart},
		End:   types.Position{Line: lineNum, Character: len(s)},
	}

	return result
}
