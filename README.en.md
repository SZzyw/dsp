# DS2API

[![License](https://img.shields.io/github/license/CJackHwang/ds2api.svg)](LICENSE)
![Stars](https://img.shields.io/github/stars/CJackHwang/ds2api.svg)
![Forks](https://img.shields.io/github/forks/CJackHwang/ds2api.svg)
[![Version](https://img.shields.io/badge/version-1.6.11-blue.svg)](version.txt)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](DEPLOY.en.md)

Language: [中文](README.MD) | [English](README.en.md)

DS2API converts DeepSeek Web chat capability into OpenAI-compatible and Claude-compatible APIs. The current repository is **Go backend only** with the existing React WebUI kept as static assets under `static/admin`.

## Implementation Boundary

- Backend: Go (`cmd/`, `api/`, `internal/`), no Python runtime
- Frontend: React admin panel (`webui/` source, static build served at runtime)
- Deployment: local run, Docker, Vercel serverless

## Key Capabilities

- OpenAI-compatible endpoints: `GET /v1/models`, `POST /v1/chat/completions`
- Claude-compatible endpoints: `GET /anthropic/v1/models`, `POST /anthropic/v1/messages`, `POST /anthropic/v1/messages/count_tokens`
- Multi-account rotation and automatic token refresh
- DeepSeek PoW solving via WASM
- Admin API: config management, account tests, import/export, Vercel sync
- WebUI SPA hosting at `/admin`
- Health probes: `GET /healthz`, `GET /readyz`

## Model Support

### OpenAI endpoint

| Model | thinking | search |
| --- | --- | --- |
| `deepseek-chat` | false | false |
| `deepseek-reasoner` | true | false |
| `deepseek-chat-search` | false | true |
| `deepseek-reasoner-search` | true | true |

### Claude endpoint

| Model | Default mapping |
| --- | --- |
| `claude-sonnet-4-20250514` | `deepseek-chat` |
| `claude-sonnet-4-20250514-fast` | `deepseek-chat` |
| `claude-sonnet-4-20250514-slow` | `deepseek-reasoner` |

You can override mapping via `claude_mapping` or `claude_model_mapping` in config.

## Quick Start

### 1) Local run

Requirement: Go 1.25+

```bash
git clone https://github.com/CJackHwang/ds2api.git
cd ds2api

cp config.example.json config.json
# edit config.json

go run ./cmd/ds2api
```

Default URL: `http://localhost:5001`

If `/admin` says WebUI not built:

```bash
./scripts/build-webui.sh
```

### 2) Docker

```bash
cp .env.example .env
# edit .env

docker-compose up -d
docker-compose logs -f
```

### 3) Vercel

- Entrypoint: `api/index.go`
- Rewrites: `vercel.json`
- Minimum env vars:
- `DS2API_ADMIN_KEY`
- `DS2API_CONFIG_JSON` (raw JSON or Base64)

## Configuration

### `config.json` example

```json
{
  "keys": ["your-api-key-1", "your-api-key-2"],
  "accounts": [
    {
      "email": "user@example.com",
      "password": "your-password",
      "token": ""
    },
    {
      "mobile": "12345678901",
      "password": "your-password",
      "token": ""
    }
  ],
  "claude_model_mapping": {
    "fast": "deepseek-chat",
    "slow": "deepseek-reasoner"
  }
}
```

### Core environment variables

| Variable | Purpose |
| --- | --- |
| `PORT` | Service port, default `5001` |
| `LOG_LEVEL` | `DEBUG/INFO/WARN/ERROR` |
| `DS2API_ADMIN_KEY` | Admin login key, default `admin` |
| `DS2API_JWT_SECRET` | Admin JWT signing secret (optional) |
| `DS2API_JWT_EXPIRE_HOURS` | Admin JWT TTL in hours, default `24` |
| `DS2API_CONFIG_PATH` | Config file path, default `config.json` |
| `DS2API_CONFIG_JSON` | Inline config (JSON or Base64) |
| `DS2API_WASM_PATH` | PoW wasm path |
| `DS2API_STATIC_ADMIN_DIR` | Admin static assets dir |
| `VERCEL_TOKEN` | Vercel sync token (optional) |
| `VERCEL_PROJECT_ID` | Vercel project ID (optional) |
| `VERCEL_TEAM_ID` | Vercel team ID (optional) |

## Auth and Account Modes

For business endpoints (`/v1/*`, `/anthropic/*`), DS2API supports two modes:

1. Managed account mode: use a key from `config.keys` via `Authorization: Bearer ...` or `x-api-key`.
2. Direct token mode: if the incoming token is not in `config.keys`, DS2API treats it as a DeepSeek token directly.

Optional header: `X-Ds2-Target-Account` to pin one managed account.

## Tool Call Adaptation

Tool-call leakage is handled in the current implementation:

- With `tools` + `stream=true`, DS2API buffers text deltas first
- If a tool call is detected, DS2API returns structured `tool_calls` only
- If no tool call is detected, DS2API emits the buffered text once
- Parser supports mixed text, fenced JSON, and `function.arguments` payloads

## Docs and Testing

- API docs: `API.md` / `API.en.md`
- Deployment docs: `DEPLOY.md` / `DEPLOY.en.md`
- Contributing: `CONTRIBUTING.md` / `CONTRIBUTING.en.md`

```bash
go test ./...
```

## Disclaimer

This project is built through reverse engineering and is provided for learning and research only. Stability is not guaranteed. Do not use it in scenarios that violate terms of service or laws.
