# End to end tests

The **/testdata** directory contains a set of subdirectories, each of which is
an end-to-end test scenario.

## Running the tests

End to end tests only run if the openapi-language-server binary is found in the
PATH.

1. Install the openapi-language-server binary:

```bash
go install
```

2. Run the tests:

```bash
go test ./internal/e2etest -count=1
```

(Use `-count=1` to disable test caching.)

## Adding or updating tests

The openapi-language-server command supports a useful flag for generating test
data:

```
-testdata string
        Capture a copy of all input and output to the specified directory. Useful for debugging or generating test data.
```

Add this flag to your editor's language server configuration to capture test
data for a specific scenario.

For example, in Neovim:

```lua
vim.lsp.start {
  cmd = { 'openapi-language-server', '-testdata', '/path/to/testdata' },
}
```

Now you can use your editor to interact with the language server and generate
test data for that scenario.
