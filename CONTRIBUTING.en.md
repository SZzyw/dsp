# Contributing Guide

Language: [中文](CONTRIBUTING.md) | [English](CONTRIBUTING.en.md)

Thanks for contributing to DS2API.

## Development Setup

### Backend (Go)

```bash
# 1. Clone
git clone https://github.com/CJackHwang/ds2api.git
cd ds2api

# 2. Configure
cp config.example.json config.json
# Edit config.json

# 3. Run backend
go run ./cmd/ds2api
```

### Frontend (WebUI)

```bash
cd webui
npm install
npm run dev
```

WebUI locales are in `webui/src/locales/`.

## Code Standards

- **Go**: run `gofmt` and make sure `go test ./...` passes before committing
- **JavaScript/React**: follow existing project style (functional components)
- **Commit messages**: use semantic prefixes (`feat:`, `fix:`, `docs:`)

## Submitting a PR

1. Fork the repo
2. Create a branch (e.g. `feature/xxx`)
3. Commit changes
4. Push your branch
5. Open a Pull Request

## Build WebUI

```bash
./scripts/build-webui.sh
```

## Project Structure

```text
ds2api/
├── cmd/ds2api/         # Local/container entrypoint
├── api/index.go        # Vercel serverless entrypoint
├── internal/           # Go backend implementation
├── webui/              # React WebUI source
├── static/admin/       # WebUI build output
├── Dockerfile
├── docker-compose.yml
└── vercel.json
```

## Reporting Issues

Please use [GitHub Issues](https://github.com/CJackHwang/ds2api/issues) with reproducible steps and logs.
