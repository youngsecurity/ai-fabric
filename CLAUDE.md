# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fabric is an open-source CLI framework for augmenting humans using AI. It provides a modular system for applying AI "Patterns" (carefully crafted prompts) to everyday tasks. The project is written in Go with a SvelteKit web interface.

## Build and Development Commands

### Go CLI

```bash
# Build and install
go install github.com/danielmiessler/fabric@latest

# Run tests (all packages)
go test -v ./...

# Run tests for a specific package
go test -v ./cli/...
go test -v ./plugins/ai/openai/...

# Check formatting with Nix
nix flake check

# Run the CLI
go run . --help
```

### Web Interface (SvelteKit)

```bash
cd web

# Install dependencies
pnpm install

# Development server
pnpm run dev

# Build for production
pnpm run build

# Run tests
pnpm run test

# Lint and format
pnpm run lint
pnpm run format

# Type checking
pnpm run check
```

### Running Fabric

```bash
# Initial setup (configures API keys, downloads patterns)
fabric --setup

# Run with a pattern
fabric --pattern summarize --stream

# Serve REST API (required for web interface)
fabric --serve

# Serve with Ollama-compatible endpoints
fabric --serveOllama
```

## Architecture

### Core Components

- **`main.go`**: Entry point, delegates to `cli.Cli()`
- **`cli/`**: Command-line interface handling
  - `cli.go`: Main CLI logic, orchestrates commands
  - `flags.go`: Flag definitions using `go-flags`
- **`core/`**: Core business logic
  - `chatter.go`: Handles chat sessions with AI vendors
  - `plugin_registry.go`: Manages all plugins (AI vendors, tools, extensions)
- **`common/`**: Shared types and utilities
  - `domain.go`: Core types like `ChatRequest`, `ChatOptions`

### Plugin System

Located in `plugins/`:

- **`ai/`**: AI vendor implementations (OpenAI, Anthropic, Ollama, etc.)
  - Each vendor implements the `Vendor` interface: `ListModels()`, `Send()`, `SendStream()`
  - `vendors.go`: `VendorsManager` coordinates multiple vendors
- **`db/fsdb/`**: Filesystem-based database for patterns, sessions, contexts
  - Data stored in `~/.config/fabric/`
- **`template/`**: Template processing for patterns with plugins (text, datetime, file, fetch, sys)
- **`strategy/`**: Prompt strategies (Chain of Thought, etc.)
- **`tools/`**: Utilities (YouTube transcripts, Jina scraping, pattern loading)

### REST API

Located in `restapi/`:
- Uses Gin framework
- Endpoints for patterns, contexts, sessions, chat, models, strategies
- Serves on `:8080` by default

### Web Interface

Located in `web/`:
- SvelteKit + Tailwind CSS + Skeleton UI
- Communicates with `fabric --serve` backend
- Key directories:
  - `src/lib/components/`: UI components
  - `src/lib/api/`: API client functions
  - `src/lib/store/`: Svelte stores for state management

### Patterns

Located in `patterns/`:
- Each pattern is a directory containing `system.md` (required) and optionally `user.md`
- Patterns use Markdown with structured sections: IDENTITY, GOALS, STEPS, OUTPUT, OUTPUT INSTRUCTIONS
- Support template variables: `{{variable}}`, `{{plugin:namespace:operation:value}}`
- Stored in `~/.config/fabric/patterns/` at runtime

### Strategies

Located in `strategies/`:
- JSON files defining prompt modifications (e.g., `cot.json` for Chain of Thought)
- Applied to system prompts before sending to LLM

## Key Interfaces

### AI Vendor Interface (`plugins/ai/vendor.go`)

```go
type Vendor interface {
    plugins.Plugin
    ListModels() ([]string, error)
    SendStream([]*goopenai.ChatCompletionMessage, *common.ChatOptions, chan string) error
    Send(context.Context, []*goopenai.ChatCompletionMessage, *common.ChatOptions) (string, error)
}
```

### Plugin Interface (`plugins/plugin.go`)

```go
type Plugin interface {
    GetName() string
    IsConfigured() bool
    Configure() error
    Setup() error
    SetupFillEnvFileContent(*bytes.Buffer)
}
```

## Configuration

- Configuration stored in `~/.config/fabric/.env`
- API keys set via environment variables or `fabric --setup`
- Common env vars: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `OLLAMA_URL`, etc.

## Adding New AI Vendors

1. Create new directory in `plugins/ai/<vendor_name>/`
2. Implement the `Vendor` interface
3. Add to `VendorsAll` in `core/plugin_registry.go`
