package main

//go:generate go run go.uber.org/mock/mockgen@v0.4.0 -source internal/lsp/handler.go -destination internal/lsp/testutil/handler.go -package testutil
