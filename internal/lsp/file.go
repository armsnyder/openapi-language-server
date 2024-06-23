package lsp

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

// File is a representation of a text file that can be modified by LSP text
// document change events. It keeps track of line breaks to allow for efficient
// conversion between byte offsets and LSP positions.
type File struct {
	bytes       []byte
	lineOffsets []int
}

// Bytes returns the raw bytes of the file.
func (f File) Bytes() []byte {
	return f.bytes
}

// Reset initializes the file with the given content.
func (f *File) Reset(s []byte) {
	f.bytes = s
	newlineCount := bytes.Count(f.bytes, []byte{'\n'})
	f.lineOffsets = make([]int, 1, newlineCount+1)

	for i, b := range f.bytes {
		if b == '\n' {
			f.lineOffsets = append(f.lineOffsets, i+1)
		}
	}
}

// ApplyChange applies the given change to the file content.
func (f *File) ApplyChange(change types.TextDocumentContentChangeEvent) error {
	if change.Range == nil {
		f.Reset([]byte(change.Text))
		return nil
	}

	start, err := f.GetOffset(change.Range.Start)
	if err != nil {
		return err
	}

	end, err := f.GetOffset(change.Range.End)
	if err != nil {
		return err
	}

	f.Reset(append(f.bytes[:start], append([]byte(change.Text), f.bytes[end:]...)...))

	return nil
}

// GetPosition returns the LSP protocol position for the given byte offset.
func (f *File) GetPosition(offset int) (types.Position, error) {
	if offset < 0 || offset > len(f.bytes) {
		return types.Position{}, fmt.Errorf("offset %d is out of range [0, %d]", offset, len(f.bytes))
	}

	line, found := slices.BinarySearch(f.lineOffsets, offset)
	if !found {
		line--
	}

	character := UTF16Len(f.bytes[f.lineOffsets[line]:offset])

	return types.Position{Line: line, Character: character}, nil
}

// GetOffset returns the byte offset for the given LSP protocol position.
func (f *File) GetOffset(p types.Position) (int, error) {
	if p.Line < 0 || p.Line >= len(f.lineOffsets) {
		return 0, fmt.Errorf("position %s is out of range", p)
	}

	if p.Line == len(f.lineOffsets) {
		if p.Character == 0 {
			return len(f.bytes), nil
		}

		return 0, fmt.Errorf("position %s is out of range", p)
	}

	rest := f.bytes[f.lineOffsets[p.Line]:]

	for i := 0; i < p.Character; i++ {
		r, size := utf8.DecodeRune(rest)

		if size == 0 || r == '\n' {
			return 0, fmt.Errorf("position %s is out of range", p)
		}

		if r == utf8.RuneError {
			return 0, errors.New("invalid UTF-8 encoding")
		}

		if r >= 0x10000 {
			// UTF-16 surrogate pair
			i++

			if i == p.Character {
				return 0, fmt.Errorf("position %s does not point to a valid UTF-16 code unit", p)
			}
		}

		rest = rest[size:]
	}

	return len(f.bytes) - len(rest), nil
}
