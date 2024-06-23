package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/armsnyder/openapiv3-lsp/internal/analysis"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

// NOTE(asanyder): Version is set by release-please.
const Version = "0.1.0"

func main() {
	// Parse command line flags.

	var args struct {
		version  bool
		help     bool
		testdata string
	}

	flag.BoolVar(&args.version, "version", false, "Print the version and exit")
	flag.BoolVar(&args.version, "v", false, "Print the version and exit")
	flag.BoolVar(&args.help, "help", false, "Print this help message and exit")
	flag.BoolVar(&args.help, "h", false, "Print this help message and exit")
	flag.StringVar(&args.testdata, "testdata", "", "Capture a copy of all input and output to the specified directory. Useful for debugging or generating test data.")

	flag.Parse()

	// Handle special flags.

	if args.version {
		//nolint:forbidigo // use of fmt.Println
		fmt.Println(Version)
		return
	}

	if args.help {
		flag.Usage()
		return
	}

	// Configure logging.

	log.SetFlags(log.Lshortfile)

	// Configure input and output.

	var reader io.Reader = os.Stdin
	var writer io.Writer = os.Stdout

	if args.testdata != "" {
		if err := os.MkdirAll(args.testdata, 0o755); err != nil {
			log.Fatal("Failed to create testdata directory: ", err)
		}

		inputFile, err := os.Create(args.testdata + "/input.jsonrpc")
		if err != nil {
			log.Fatal("Failed to create input file: ", err)
		}
		defer inputFile.Close()

		outputFile, err := os.Create(args.testdata + "/output.jsonrpc")
		if err != nil {
			//nolint:gocritic // exitAfterDefer
			log.Fatal("Failed to create output file: ", err)
		}
		defer outputFile.Close()

		reader = io.TeeReader(reader, inputFile)
		writer = io.MultiWriter(writer, outputFile)
	}

	// Run the LSP server.

	server := &lsp.Server{
		ServerInfo: types.ServerInfo{
			Name:    "openapiv3-lsp",
			Version: Version,
		},
		Reader:  reader,
		Writer:  writer,
		Handler: &analysis.Handler{},
	}

	if err := server.Run(); err != nil {
		log.Fatal("LSP server error: ", err)
	}
}

// shadowReader wraps a primary reader and splits the input to a secondary
// writer.
type shadowReader struct {
	reader io.Reader
	shadow io.Writer
}

func (r shadowReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	_, _ = r.shadow.Write(p[:n])
	return n, err
}

var _ io.Reader = shadowReader{}

// shadowWriter wraps a primary writer and splits the output to a secondary
// writer.
type shadowWriter struct {
	writer io.Writer
	shadow io.Writer
}

func (w shadowWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	_, _ = w.shadow.Write(p)
	return n, err
}

var _ io.Writer = shadowWriter{}
