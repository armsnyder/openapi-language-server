package lsp_test

import (
	"regexp"
	"runtime"
	"strconv"
	"testing"

	. "github.com/armsnyder/openapi-language-server/internal/lsp"
	"github.com/armsnyder/openapi-language-server/internal/lsp/types"
)

func TestFile(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		steps []Step
	}{
		{
			name: "hand type from empty",
			text: "\n",
			steps: []Step{
				{Text: "a", Range: "0:0-0:0", Want: "a\n"},
				{Text: "b", Range: "0:1-0:1", Want: "ab\n"},
				{Text: "\n", Range: "0:2-0:2", Want: "ab\n\n"},
				{Text: "c", Range: "1:0-1:0", Want: "ab\nc\n"},
				{Text: "d", Range: "1:1-1:1", Want: "ab\ncd\n"},
			},
		},
		{
			name: "hand delete from end",
			text: "ab\ncd\n",
			steps: []Step{
				{Text: "", Range: "1:1-1:2", Want: "ab\nc\n"},
				{Text: "", Range: "1:0-1:1", Want: "ab\n\n"},
				{Text: "", Range: "0:2-0:2", Want: "ab\n\n"},
				{Text: "", Range: "1:0-2:0", Want: "ab\n"},
				{Text: "", Range: "0:1-0:2", Want: "a\n"},
				{Text: "", Range: "0:0-0:1", Want: "\n"},
			},
		},
		{
			name: "add 2 lines to the middle then update each line",
			text: "ab\ncd\n",
			steps: []Step{
				{Text: "\n12\n34", Range: "0:2-0:2", Want: "ab\n12\n34\ncd\n"},
				{Text: "x", Range: "3:1-3:2", Want: "ab\n12\n34\ncx\n"},
				{Text: "y", Range: "2:1-2:2", Want: "ab\n12\n3y\ncx\n"},
				{Text: "z", Range: "1:1-1:2", Want: "ab\n1z\n3y\ncx\n"},
			},
		},
		{
			name: "insert text at the beginning of the file",
			text: "line1\nline2\nline3\n",
			steps: []Step{
				{Text: "start\n", Range: "0:0-0:0", Want: "start\nline1\nline2\nline3\n"},
			},
		},
		{
			name: "insert text at the end of the file",
			text: "line1\nline2\nline3\n",
			steps: []Step{
				{Text: "end\n", Range: "3:0-3:0", Want: "line1\nline2\nline3\nend\n"},
			},
		},
		{
			name: "insert newline at the beginning and end of the file",
			text: "line1\nline2\nline3\n",
			steps: []Step{
				{Text: "\n", Range: "0:0-0:0", Want: "\nline1\nline2\nline3\n"},
				{Text: "\n", Range: "4:0-4:0", Want: "\nline1\nline2\nline3\n\n"},
			},
		},
		{
			name: "delete text spanning multiple lines",
			text: "line1\nline2\nline3\nline4\n",
			steps: []Step{
				{Text: "", Range: "1:2-3:4", Want: "line1\nli4\n"},
				{Text: "x", Range: "1:3-1:3", Want: "line1\nli4x\n"},
			},
		},
		{
			name: "replace text spanning multiple lines with text containing newlines",
			text: "line1\nline2\nline3\nline4\n",
			steps: []Step{
				{Text: "new\ntext\n", Range: "1:2-3:4", Want: "line1\nlinew\ntext\n4\n"},
				{Text: "x", Range: "3:1-3:1", Want: "line1\nlinew\ntext\n4x\n"},
			},
		},
		{
			name: "delete final line",
			text: "a\n",
			steps: []Step{
				{Text: "", Range: "0:0-1:0", Want: ""},
			},
		},
		{
			name: "add to empty file without newline",
			text: "",
			steps: []Step{
				{Text: "\n\n", Range: "0:0-0:0", Want: "\n\n"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f File
			f.Reset([]byte(tt.text))
			for i, step := range tt.steps {
				func() {
					defer func() {
						if r := recover(); r != nil {
							stack := make([]byte, 1<<16)
							stack = stack[:runtime.Stack(stack, false)]
							t.Fatalf("step %d: %v\n%s", i, r, stack)
						}
					}()
					event := newEvent(step.Text, step.Range)
					if err := f.ApplyChange(event); err != nil {
						t.Fatalf("step %d: %v", i, err)
					}
					if got := string(f.Bytes()); got != step.Want {
						t.Fatalf("step %d: got %q, want %q", i, got, step.Want)
					}
				}()
			}
		})
	}
}

type Step struct {
	Text  string
	Range string
	Want  string
}

var rangePattern = regexp.MustCompile(`^(\d+):(\d+)-(\d+):(\d+)$`)

func newEvent(text, rng string) types.TextDocumentContentChangeEvent {
	match := rangePattern.FindSubmatch([]byte(rng))
	if match == nil {
		panic("invalid range")
	}

	return types.TextDocumentContentChangeEvent{
		Text: text,
		Range: &types.Range{
			Start: types.Position{
				Line:      mustAtoi(match[1]),
				Character: mustAtoi(match[2]),
			},
			End: types.Position{
				Line:      mustAtoi(match[3]),
				Character: mustAtoi(match[4]),
			},
		},
	}
}

func mustAtoi(b []byte) int {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		panic(err)
	}
	return i
}
