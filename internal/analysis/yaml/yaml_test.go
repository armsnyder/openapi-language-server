package yaml_test

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	. "github.com/armsnyder/openapi-language-server/internal/analysis/yaml"
	"github.com/armsnyder/openapi-language-server/internal/lsp/types"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want Line
	}{
		{
			name: "empty",
			line: "",
			want: Line{},
		},
		{
			name: "key only",
			line: "foo:",
			want: Line{
				Key: "foo",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 3},
				},
			},
		},
		{
			name: "key and value",
			line: "foo: bar",
			want: Line{
				Key:   "foo",
				Value: "bar",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 3},
				},
				ValueRange: types.Range{
					Start: types.Position{Line: 0, Character: 5},
					End:   types.Position{Line: 0, Character: 8},
				},
			},
		},
		{
			name: "key and value with leading whitespace",
			line: "  foo: bar",
			want: Line{
				Key:   "foo",
				Value: "bar",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 2},
					End:   types.Position{Line: 0, Character: 5},
				},
				ValueRange: types.Range{
					Start: types.Position{Line: 0, Character: 7},
					End:   types.Position{Line: 0, Character: 10},
				},
			},
		},
		{
			name: "double quoted value",
			line: `foo: "bar"`,
			want: Line{
				Key:   "foo",
				Value: "bar",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 3},
				},
				ValueRange: types.Range{
					Start: types.Position{Line: 0, Character: 6},
					End:   types.Position{Line: 0, Character: 9},
				},
			},
		},
		{
			name: "single quoted value",
			line: `foo: 'bar'`,
			want: Line{
				Key:   "foo",
				Value: "bar",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 3},
				},
				ValueRange: types.Range{
					Start: types.Position{Line: 0, Character: 6},
					End:   types.Position{Line: 0, Character: 9},
				},
			},
		},
		{
			name: "extra space before value",
			line: "foo:  bar",
			want: Line{
				Key:   "foo",
				Value: "bar",
				KeyRange: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 3},
				},
				ValueRange: types.Range{
					Start: types.Position{Line: 0, Character: 6},
					End:   types.Position{Line: 0, Character: 9},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document, err := Parse(bytes.NewReader([]byte(tt.line + "\n")))
			if err != nil {
				t.Fatal(err)
			}
			if len(document.Lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(document.Lines))
			}
			line := document.Lines[0]
			if line.Key != tt.want.Key {
				t.Errorf("got key %q, want %q", line.Key, tt.want.Key)
			}
			if line.Value != tt.want.Value {
				t.Errorf("got value %q, want %q", line.Value, tt.want.Value)
			}
			if len(line.Children) > 0 {
				t.Errorf("got %d children, want 0", len(line.Children))
			}
			if line.Parent != nil {
				t.Errorf("got parent, want nil")
			}
			if line.KeyRange != tt.want.KeyRange {
				t.Errorf("got key range %v, want %v", line.KeyRange, tt.want.KeyRange)
			}
			if line.ValueRange != tt.want.ValueRange {
				t.Errorf("got value range %v, want %v", line.ValueRange, tt.want.ValueRange)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		document, err := Parse(bytes.NewReader([]byte("")))
		if err != nil {
			t.Fatal(err)
		}

		if len(document.Lines) != 0 {
			t.Errorf("got %d lines, want 0", len(document.Lines))
		}

		if len(document.Root) != 0 {
			t.Errorf("got %d root keys, want 0", len(document.Root))
		}
	})

	t.Run("one line", func(t *testing.T) {
		document, err := Parse(bytes.NewReader([]byte("foo: bar\n")))
		if err != nil {
			t.Fatal(err)
		}

		if len(document.Lines) != 1 {
			t.Errorf("got %d lines, want 1", len(document.Lines))
		}

		if len(document.Root) != 1 {
			t.Errorf("got %d root keys, want 1", len(document.Root))
		}

		line := document.Lines[0]
		if line.Key != "foo" {
			t.Errorf("got key %q, want %q", line.Key, "foo")
		}

		if line.Value != "bar" {
			t.Errorf("got value %q, want %q", line.Value, "bar")
		}

		if len(line.Children) != 0 {
			t.Errorf("got %d children, want 0", len(line.Children))
		}

		if line.Parent != nil {
			t.Errorf("got parent, want nil")
		}

		if document.Root["foo"] != line {
			t.Errorf("root key and line do not match")
		}
	})

	t.Run("nested", func(t *testing.T) {
		document, err := Parse(bytes.NewReader([]byte("foo:\n  bar: baz\n")))
		if err != nil {
			t.Fatal(err)
		}

		if len(document.Lines) != 2 {
			t.Errorf("got %d lines, want 2", len(document.Lines))
		}

		if len(document.Root) != 1 {
			t.Errorf("got %d root keys, want 1", len(document.Root))
		}

		foo := document.Root["foo"]
		if foo == nil {
			t.Fatal("missing root key foo")
		}

		if foo.Key != "foo" {
			t.Errorf("foo: got key %q, want %q", foo.Key, "foo")
		}

		if foo.Value != "" {
			t.Errorf("foo: got value %q, want %q", foo.Value, "")
		}

		if len(foo.Children) != 1 {
			t.Errorf("foo: got %d children, want 1", len(foo.Children))
		}

		if foo.Parent != nil {
			t.Errorf("foo: got parent, want nil")
		}

		bar := foo.Children["bar"]
		if bar == nil {
			t.Fatal("foo: missing child key bar")
		}

		if bar.Key != "bar" {
			t.Errorf("bar: got key %q, want %q", bar.Key, "bar")
		}

		if bar.Value != "baz" {
			t.Errorf("bar: got value %q, want %q", bar.Value, "baz")
		}

		if len(bar.Children) != 0 {
			t.Errorf("bar: got %d children, want 0", len(bar.Children))
		}

		if bar.Parent != foo {
			t.Errorf("bar: parent mismatch")
		}

		if foo != document.Lines[0] {
			t.Errorf("foo line does not match")
		}

		if bar != document.Lines[1] {
			t.Errorf("bar line does not match")
		}
	})
}

func TestRefs(t *testing.T) {
	document, err := Parse(bytes.NewReader([]byte(`
# comment
openapi: 3.0.0
paths:
  /foo:
    get:
      $ref: "#/components/schemas/Foo"
components:
  schemas:
    Foo:
      type: object
    Bar:
      type: object
`)))
	if err != nil {
		t.Fatal(err)
	}

	const undefined = "undefined"

	expectedRefs := []string{
		undefined,
		undefined,
		"#/openapi",
		"#/paths",
		undefined,
		undefined,
		undefined,
		"#/components",
		"#/components/schemas",
		"#/components/schemas/Foo",
		"#/components/schemas/Foo/type",
		"#/components/schemas/Bar",
		"#/components/schemas/Bar/type",
	}

	for i, expectedRef := range expectedRefs {
		if expectedRef == undefined {
			continue
		}

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			line := document.Lines[i]
			gotRef := line.KeyRef()
			if gotRef != expectedRef {
				t.Errorf("KeyRef: got %q, want %q", gotRef, expectedRef)
			}

			if expectedRef != "" {
				loc := document.Locate(expectedRef)
				if loc != line {
					t.Errorf("Locate: got %v, want %v", loc, line)
				}
			}
		})
	}
}

func TestParse_PetStore(t *testing.T) {
	f, err := os.Open("testdata/petstore.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	document, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	line := document.Locate("#/components/schemas/Pet")
	if line == nil {
		t.Fatal("could not locate Pet schema")
	}

	wantRange := types.Range{
		Start: types.Position{Line: 719, Character: 4},
		End:   types.Position{Line: 719, Character: 7},
	}
	if line.KeyRange != wantRange {
		t.Errorf("got key range %v, want %v", line.KeyRange, wantRange)
	}

	wantKey := "Pet"
	if line.Key != wantKey {
		t.Errorf("got key %q, want %q", line.Key, wantKey)
	}
}
