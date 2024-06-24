# OpenAPI Language Server

An OpenAPI language server for [LSP compatible code
editors.](https://microsoft.github.io/language-server-protocol/implementors/tools/)

> :warning: This is beta software. Many features are still missing. See
> [Features](https://github.com/armsnyder/openapi-language-server?tab=readme-ov-file#features)
> below.

[![asciicast](https://asciinema.org/a/v7etZb80HbYkKBQUa3dVSenPz.svg)](https://asciinema.org/a/v7etZb80HbYkKBQUa3dVSenPz)

## Features

I created this language server because I manually edit OpenAPI/Swagger files,
and I needed a quick way to jump between schema refinitions and references.

I use
[yaml-language-server](https://github.com/redhat-developer/yaml-language-server)
for validation and completion, so these features are not a priority for me
right now.

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
- [ ]

## Installation

### From GitHub Releases (Recommended)

Download the latest release from [GitHub releases](https://github.com/armsnyder/openapi-language-server/releases).

### Using Go

```bash
go install github.com/armsnyder/openapi-language-server@latest
```

## Usage

### Neovim Configuration Example

Assuming you are using Neovim and have the installed openapi-language-server
binary in your PATH, you can use the following Lua code in your Neovim
configuration:

```lua
    vim.api.nvim_create_autocmd('FileType', {
      pattern = 'yaml',
      callback = function()
        vim.lsp.start {
          cmd = { 'openapi-language-server' },
          filetypes = { 'yaml' },
          root_dir = vim.fn.getcwd(),
        }
      end,
    })
```

This is just a basic working example. You will probably want to further
customize the configuration to your needs.
