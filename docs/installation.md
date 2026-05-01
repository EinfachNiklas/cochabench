# Installation

## Installation Methods

### Install via Go (Recommended)

Before installing CochaBench via Go, ensure you have the following:

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.25.5+ | Install Cochabench |

The simplest way to install CochaBench globally:

```bash
go install github.com/EinfachNiklas/cochabench/cmd/cochabench@latest
```

This installs the `cochabench` binary to your `$GOPATH/bin` directory.

---
### Build from Source

Before building Cochabench from source, ensure you have the following:

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.25.5+ | Build the application |
| Git | Any recent | Clone repository to your device |

For development or to build a specific version:

```bash
# Clone the repository
git clone https://github.com/EinfachNiklas/cochabench.git
cd cochabench

# Build the binary
go build -o cochabench ./cmd/cochabench

# Optionally move to your PATH
mv cochabench /usr/local/bin/
```

---
## Verify Installation

After installation, verify CochaBench is working:

```bash
cochabench --version
```

---
## Language-Specific Requirements

### JavaScript Challenges

JavaScript challenges require Node.js and npm:

```bash
# Verify npm is available
npm --version
```

CochaBench runs `npm install` and `npm test` during evaluation.

### Python Challenges

Python challenges require Python 3 with venv support:

```bash
# Verify Python is available
python3 --version
# or
python --version

# Verify venv module is available
python3 -m venv --help
```

CochaBench creates a virtual environment and installs dependencies via `pip`.

### Go Challenges

Go challenges use the Go toolchain:

```bash
# Verify Go is available
go version
```

CochaBench runs `go mod download`, `go mod tidy`, and `go test` during evaluation.

---
## AI Evaluation Setup

To enable AI-powered code quality evaluation, you need an API key from a supported provider.

### Supported Providers

| Provider | Environment Variable |
|----------|---------------------|
| Anthropic (Claude) | `LLM_API_KEY` |
| OpenAI (GPT) | `LLM_API_KEY` |
| Google (Gemini) | `LLM_API_KEY` |

### Set Your API Key

```bash
# Linux/macOS
export LLM_API_KEY="your-api-key-here"

# Windows (PowerShell)
$env:LLM_API_KEY="your-api-key-here"

# Windows (CMD)
set LLM_API_KEY=your-api-key-here
```

For persistent configuration, add the export to your shell profile (`.bashrc`, `.zshrc`, etc.).

---
## Optional: GitHub Token

If connecting to a private challenge server repository, set a GitHub personal access token and update the configuration to use the new url of the private server:

```bash
export GITHUB_TOKEN="your-github-token"
cochabench config set CHALLENGE_SERVER "your-server-url"
```

> [!NOTE]
> See also [Configuration - Challenge Server](configuration.md#challenge_server)

---
## Next Steps

- [Initialize your configuration](configuration.md)
- [Complete your first challenge](quickstart.md)
