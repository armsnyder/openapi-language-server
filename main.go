package main

import (
	"log"
	"os"

	"github.com/armsnyder/openapiv3-lsp/internal/analysis"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp"
	"github.com/armsnyder/openapiv3-lsp/internal/lsp/types"
)

func main() {
	log.SetFlags(0)

	server := &lsp.Server{
		ServerInfo: types.ServerInfo{
			Name:    "openapiv3-lsp",
			Version: "0.1.0",
		},
		Reader:  os.Stdin,
		Writer:  os.Stdout,
		Handler: &analysis.Handler{},
	}

	if err := server.Run(); err != nil {
		log.Fatal("LSP server error: ", err)
	}
}
