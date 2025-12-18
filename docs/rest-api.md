# Fabric REST API

Fabric's REST API provides HTTP access to all core functionality: chat completions, pattern management, contexts, sessions, and more.

## Quick Start

Start the server:

```bash
fabric --serve
```

The server runs on `http://localhost:8080` by default.

Test it:

```bash
curl http://localhost:8080/patterns/names
```

## Server Options

| Flag | Description | Default |
|------|-------------|---------|
| `--serve` | Start the REST API server | - |
| `--address` | Server address and port | `:8080` |
| `--api-key` | Enable API key authentication | (none) |

Example with custom configuration:

```bash
fabric --serve --address :9090 --api-key my_secret_key
```

## Authentication

When you set an API key with `--api-key`, all requests must include:

```http
X-API-Key: your-api-key-here
```

Example:

```bash
curl -H "X-API-Key: my_secret_key" http://localhost:8080/patterns/names
```

Without an API key, the server accepts all requests and logs a warning.

## Endpoints

### Chat Completions

Stream AI responses using Server-Sent Events (SSE).

**Endpoint:** `POST /chat`

**Request:**

```json
{
  "prompts": [
    {
      "userInput": "Explain quantum computing",
      "vendor": "openai",
      "model": "gpt-4o",
      "patternName": "explain",
      "contextName": "",
      "strategyName": "",
      "variables": {}
    }
  ],
  "language": "en",
  "temperature": 0.7,
  "topP": 0.9,
  "frequencyPenalty": 0,
  "presencePenalty": 0,
  "thinking": 0
}
```

**Prompt Fields:**

- `userInput` - Your message or question
- `vendor` - AI provider: `openai`, `anthropic`, `gemini`, `ollama`, etc.
- `model` - Model name: `gpt-4o`, `claude-sonnet-4.5`, `gemini-2.0-flash-exp`, etc.
- `patternName` - Pattern to apply (optional, from `~/.config/fabric/patterns/`)
- `contextName` - Context to prepend (optional, from `~/.config/fabric/contexts/`)
- `strategyName` - Strategy to use (optional, from `~/.config/fabric/strategies/`)
- `variables` - Variable substitutions for patterns (optional, e.g., `{"role": "expert"}`)

**Response:**

Server-Sent Events stream with `Content-Type: text/readystream`. Each line contains JSON:

```json
{"type": "content", "format": "markdown", "content": "Quantum computing uses..."}
{"type": "content", "format": "markdown", "content": " quantum mechanics..."}
{"type": "complete", "format": "markdown", "content": ""}
```

**Types:**

- `content` - Response chunk
- `error` - Error message
- `complete` - Stream finished

**Formats:**

- `markdown` - Standard text
- `mermaid` - Mermaid diagram
- `plain` - Plain text

**Example:**

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompts": [{
      "userInput": "What is Fabric?",
      "vendor": "openai",
      "model": "gpt-4o",
      "patternName": "explain"
    }]
  }'
```

### Patterns

Manage reusable AI prompts.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/patterns/names` | List all pattern names |
| `GET` | `/patterns/:name` | Get pattern content |
| `GET` | `/patterns/exists/:name` | Check if pattern exists |
| `POST` | `/patterns/:name` | Create or update pattern |
| `DELETE` | `/patterns/:name` | Delete pattern |
| `PUT` | `/patterns/rename/:oldName/:newName` | Rename pattern |
| `POST` | `/patterns/:name/apply` | Apply pattern with variables |

**Example - Get pattern:**

```bash
curl http://localhost:8080/patterns/summarize
```

**Example - Apply pattern with variables:**

```bash
curl -X POST http://localhost:8080/patterns/translate/apply \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello world",
    "variables": {"lang_code": "es"}
  }'
```

**Example - Create pattern:**

```bash
curl -X POST http://localhost:8080/patterns/my_custom_pattern \
  -H "Content-Type: text/plain" \
  -d "You are an expert in explaining complex topics simply..."
```

### Contexts

Manage context snippets that prepend to prompts.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/contexts/names` | List all context names |
| `GET` | `/contexts/:name` | Get context content |
| `GET` | `/contexts/exists/:name` | Check if context exists |
| `POST` | `/contexts/:name` | Create or update context |
| `DELETE` | `/contexts/:name` | Delete context |
| `PUT` | `/contexts/rename/:oldName/:newName` | Rename context |

### Sessions

Manage chat conversation history.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/sessions/names` | List all session names |
| `GET` | `/sessions/:name` | Get session messages (JSON array) |
| `GET` | `/sessions/exists/:name` | Check if session exists |
| `POST` | `/sessions/:name` | Save session messages |
| `DELETE` | `/sessions/:name` | Delete session |
| `PUT` | `/sessions/rename/:oldName/:newName` | Rename session |

### Models

List available AI models.

**Endpoint:** `GET /models/names`

**Response:**

```json
{
  "models": ["gpt-4o", "gpt-4o-mini", "claude-sonnet-4.5", "gemini-2.0-flash-exp"],
  "vendors": {
    "openai": ["gpt-4o", "gpt-4o-mini"],
    "anthropic": ["claude-sonnet-4.5", "claude-opus-4.5"],
    "gemini": ["gemini-2.0-flash-exp", "gemini-2.0-flash-thinking-exp"]
  }
}
```

### Strategies

List available prompt strategies (Chain of Thought, etc.).

**Endpoint:** `GET /strategies`

**Response:**

```json
[
  {
    "name": "chain_of_thought",
    "description": "Think step by step",
    "prompt": "Let's think through this step by step..."
  }
]
```

### YouTube Transcripts

Extract transcripts from YouTube videos.

**Endpoint:** `POST /youtube/transcript`

**Request:**

```json
{
  "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
  "timestamps": false
}
```

**Response:**

```json
{
  "title": "Video Title",
  "transcript": "Full transcript text..."
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/youtube/transcript \
  -H "Content-Type: application/json" \
  -d '{"url": "https://youtube.com/watch?v=dQw4w9WgXcQ", "timestamps": true}'
```

### Configuration

Manage API keys and environment settings.

**Get configuration:**

`GET /config`

Returns API keys and URLs for all configured vendors.

**Update configuration:**

`POST /config/update`

```json
{
  "OPENAI_API_KEY": "sk-...",
  "ANTHROPIC_API_KEY": "sk-ant-..."
}
```

Updates `~/.config/fabric/.env` with new values.

## Docker Usage

Run the server in Docker:

```bash
# Setup (first time)
mkdir -p $HOME/.fabric-config
docker run --rm -it \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --setup

# Start server
docker run --rm -it \
  -p 8080:8080 \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --serve

# With authentication
docker run --rm -it \
  -p 8080:8080 \
  -v $HOME/.fabric-config:/root/.config/fabric \
  kayvan/fabric:latest --serve --api-key my_secret_key
```

## Ollama Compatibility Mode

Fabric can emulate Ollama's API endpoints:

```bash
fabric --serveOllama --address :11434
```

This mode provides:

- `GET /api/tags` - Lists patterns as models
- `GET /api/version` - Server version
- `POST /api/chat` - Ollama-compatible chat endpoint

## Error Handling

All endpoints return standard HTTP status codes:

- `200 OK` - Success
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing or invalid API key
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses include JSON with details:

```json
{
  "error": "Pattern not found: nonexistent"
}
```

## Rate Limiting

The server does not implement rate limiting. When deploying publicly, use a reverse proxy (nginx, Caddy) with rate limiting enabled.

## CORS

The server sets CORS headers for local development:

```http
Access-Control-Allow-Origin: http://localhost:5173
```

For production, configure CORS through a reverse proxy.
