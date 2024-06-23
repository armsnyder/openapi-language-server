# OpenAPI Language Server

This is a language server for OpenAPI v3. It is based on the [Language
Server Protocol](https://microsoft.github.io/language-server-protocol/).

[![asciicast](https://asciinema.org/a/v7etZb80HbYkKBQUa3dVSenPz.svg)](https://asciinema.org/a/v7etZb80HbYkKBQUa3dVSenPz)

I created this language server because I do a lot of manual OpenAPI/Swagger
file editing, and I wanted a quick way to jump to definitions and find
references of schema definitions.

I personally use
[yaml-language-server](https://github.com/redhat-developer/yaml-language-server)
for schema validation and code completion, so these features are not a priority
for me to implement in this language server.

## Features

### Language Features

- [x] Jump to definition
- [x] Find references
- [ ] Code completion
- [ ] Diagnostics
- [ ] Hover
- [ ] Rename
- [ ] Document symbols
- [ ] Code actions

### Other Features

- [x] YAML filetype support
- [ ] JSON filetype support
- [ ] VSCode extension

## Installation

### Using Go

```bash
go install github.com/armsnyder/openapiv3-lsp@latest
```

### From GitHub Releases

Download the latest release from [GitHub releases](https://github.com/armsnyder/openapiv3-lsp/releases).

## Usage

### Neovim Configuration Example

Assuming you are using Neovim and have the installed openapiv3-lsp binary in
your PATH, you can use the following Lua code to your Neovim configuration:

```lua
    vim.api.nvim_create_autocmd('FileType', {
      pattern = 'yaml',
      callback = function()
        vim.lsp.start {
          cmd = { 'openapiv3-lsp' },
          filetypes = { 'yaml' },
          root_dir = vim.fn.getcwd(),
        }
      end,
    })
```

This is just a basic working example. You will probably want to further
customize the configuration to your needs.
