package analysis_test

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "github.com/armsnyder/openapiv3-lsp/internal/analysis"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

type HandlerSetupFunc func(t *testing.T, h *Handler)

func TestHandler_HandleDefinition(t *testing.T) {
	tests := []struct {
		name    string
		setup   HandlerSetupFunc
		params  types.DefinitionParams
		want    []types.Location
		wantErr bool
	}{
		{
			name:   "file not found",
			setup:  loadFile("file:///foo", "foo"),
			params: definitionParams("file:///bar", "0:0"),
		},
		{
			name:   "no definition",
			setup:  loadFile("file:///foo", "foo"),
			params: definitionParams("file:///foo", "0:0"),
		},
		{
			name: "start of ref",
			setup: loadFile("file:///foo", `
foo:
  $ref: "#/bar/baz"
bar:
  baz:
	type: object`),
			params: definitionParams("file:///foo", "2:8"),
			want:   locations("file:///foo", "4:2-4:5"),
		},
		{
			name: "end of ref",
			setup: loadFile("file:///foo", `
foo:
  $ref: "#/bar/baz"
bar:
  baz:
	type: object`),
			params: definitionParams("file:///foo", "2:18"),
			want:   locations("file:///foo", "4:2-4:5"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h Handler

			tt.setup(t, &h)

			got, err := h.HandleDefinition(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got == nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleDefinition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_HandleReferences(t *testing.T) {
	tests := []struct {
		name    string
		setup   HandlerSetupFunc
		params  types.ReferenceParams
		want    []types.Location
		wantErr bool
	}{
		{
			name:   "file not found",
			setup:  loadFile("file:///foo", "foo"),
			params: referenceParams("file:///bar", "0:0"),
		},
		{
			name:   "no references",
			setup:  loadFile("file:///foo", "foo"),
			params: referenceParams("file:///foo", "0:0"),
		},
		{
			name: "simple",
			setup: loadFile("file:///foo", `
foo:
  $ref: "#/bar/baz"
bar:
  baz:
    type: object`),
			params: referenceParams("file:///foo", "4:2"),
			want:   locations("file:///foo", "2:9-2:18"),
		},
		{
			name: "multiple references",
			setup: loadFile("file:///foo", `
foo:
  $ref: "#/bar/baz"
foo2:
  $ref: "#/bar/baz"
bar:
  baz:
	type: object`),
			params: referenceParams("file:///foo", "6:2"),
			want:   locations("file:///foo", "2:9-2:18", "4:9-4:18"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h Handler

			tt.setup(t, &h)

			got, err := h.HandleReferences(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleReferences() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.want) == 0 && len(got) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func loadFile(uri, text string) HandlerSetupFunc {
	return func(t *testing.T, h *Handler) {
		if err := h.HandleOpen(types.DidOpenTextDocumentParams{
			TextDocument: types.TextDocumentItem{
				URI:  uri,
				Text: text,
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func referenceParams(uri, position string) types.ReferenceParams {
	return types.ReferenceParams{
		TextDocumentPositionParams: positionParams(uri, position),
	}
}

func definitionParams(uri, position string) types.DefinitionParams {
	return types.DefinitionParams{
		TextDocumentPositionParams: positionParams(uri, position),
	}
}

func positionParams(uri, position string) types.TextDocumentPositionParams {
	split := strings.Split(position, ":")

	line, err := strconv.Atoi(split[0])
	if err != nil {
		panic(err)
	}

	character, err := strconv.Atoi(split[1])
	if err != nil {
		panic(err)
	}

	return types.TextDocumentPositionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: uri,
		},
		Position: types.Position{
			Line:      line,
			Character: character,
		},
	}
}

func locations(uri string, ranges ...string) []types.Location {
	pat := regexp.MustCompile(`^(\d+):(\d+)-(\d+):(\d+)$`)
	locs := make([]types.Location, len(ranges))
	for i, rng := range ranges {
		match := pat.FindStringSubmatch(rng)
		if match == nil {
			panic("invalid range")
		}

		locs[i] = types.Location{
			URI: uri,
			Range: types.Range{
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

	return locs
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
