# CochaBench

![GitHub Actions](https://github.com/EinfachNiklas/cochabench/workflows/Run%20Tests/badge.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/EinfachNiklas/cochabench)
[![Go Reference](https://pkg.go.dev/badge/github.com/EinfachNiklas/cochabench/.svg)](https://pkg.go.dev/github.com/EinfachNiklas/cochabench/)

**CochaBench** is a comprehensive coding challenge benchmark suite designed to evaluate and compare the performance of developers and AI coding agents across multiple programming languages.


## Features

- **Multi-Language Support**: Challenges available in JavaScript, Python, and Go
- **Standardized Evaluation**: Consistent metrics across all challenges including:
  - Test execution results (pass/fail rates)
  - Execution time tracking
  - AI-powered code quality assessment (quality, maintainability, security)
- **Flexible Workflow**: Initialize, start, stop, and evaluate coding challenge attempts
- **Challenge Management**: Easy download and management of challenge sets
- **Persistent Storage**: SQLite-based tracking of all runs and evaluations

## Installation

### Prerequisites

- Go 1.25.5 or higher
- Git
- For JavaScript challenge evaluation: `npm`
- For Python challenge evaluation: `python3` or `python`, `venv`, and `pip`
- For AI-assisted evaluation: an API key in `LLM_API_KEY`

### Build from Source

```bash
git clone https://github.com/EinfachNiklas/cochabench.git
cd cochabench
go build -o cochabench ./cmd/cochabench
```

### Install Globally

```bash
go install github.com/EinfachNiklas/cochabench/cmd/cochabench@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
cochabench config init
```
The config file is created at `~/.config/cochabench/config.json`

### 2. List Available Challenges

```bash
cochabench challenge list
```

### 3. Download a Challenge

```bash
cochabench challenge get <challenge-id>
```

Or download all challenges:

```bash
cochabench challenge get all
```

### 4. Create a Run

Navigate to a challenge directory and initialize a run:

```bash
cd <challenge-directory>
cochabench run init --name "my-first-attempt"
```

Each challenge directory is expected to contain:

- `src/` with the starter implementation
- `test/` with the benchmark tests
- `challenge.config.json` with challenge metadata

### 5. Start Working

```bash
cochabench run start --id <run-id>
```

Work on your solution in the `solutions/<run-id>/` directory.
CochaBench also creates a local `cochabench.db` SQLite database in the challenge directory to persist run metadata and evaluation results.

### 6. Stop the Run

```bash
cochabench run stop --id <run-id>
```

### 7. Evaluate Your Solution

```bash
cochabench run eval --runID <run-id>
```

## Usage

### Challenge Management

```bash
# List all available challenges
cochabench challenge list

# Download a specific challenge
cochabench challenge get <challenge-id>

# Download all challenges
cochabench challenge get all
```

### Run Management

```bash
# Initialize a new run
cochabench run init --name "attempt-1"

# Start a run
cochabench run start --id <run-id>

# Stop a run
cochabench run stop --id <run-id>

# Cancel a run
cochabench run cancel --id <run-id>

# List all runs for current challenge
cochabench run list
```

### Evaluation

```bash
# Evaluate a completed run
cochabench run eval --runID <run-id>

# Evaluate without AI assessment
cochabench run eval --runID <run-id> --no-ai-eval

# Debug mode (keep temporary files)
cochabench run eval --runID <run-id> --debug
```

### Configuration

```bash
# Initialize config
cochabench config init

# Show all config values
cochabench config show

# Get a specific config value
cochabench config get <key>

# Set a config value
cochabench config set <key> <value>
```

## Evaluation Metrics

CochaBench provides comprehensive evaluation metrics:

- **Test Results**: Total, passed, and failed test counts
- **Execution Time**: Duration of test execution
- **Quality Score**: AI-evaluated code quality (1-10, averaged across multiple runs)
- **Maintainability Score**: AI-evaluated code maintainability (1-10, averaged across multiple runs)
- **Security Score**: AI-evaluated code security (1-10, averaged across multiple runs)

### AI support

Currently, only Anthropic (Claude), OpenAI (ChatGPT, Codex) and Google (Gemini) are supported.

The generated config file exposes the following keys:

- `LLM_PROVIDER` (`anthropic`, `openai`, `google`)
- `LLM_BASE_PATH`
- `LLM_MODEL`
- `CHALLENGE_SERVER`

## Environment Variables

CochaBench supports the following environment variables:

- `GITHUB_TOKEN`: GitHub personal access token for API requests (optional, required to connect to private challenge server repos)
- `LLM_API_KEY`: API key used for AI evaluation

## Security Notice

CochaBench executes challenge code, test suites, and package installation commands locally on your machine.

- Do not run untrusted challenges or untrusted solution code without additional isolation.
- JavaScript evaluation runs `npm install` and `npm test`.
- Python evaluation creates a virtual environment and installs packages with `pip`.
- Go evaluation may download modules with `go mod download` and `go mod tidy`.

CochaBench does not currently provide sandboxing or container isolation for evaluation.

For vulnerability reporting and additional security guidance, see [SECURITY.md](SECURITY.md).

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
cochabench/
├── cmd/
│   └── cochabench/          # Main CLI application
├── internal/
│   ├── challenge/           # Challenge download and management
│   ├── cochabenchData/      # Database operations
│   ├── config/              # Configuration management
│   ├── eval/                # Evaluation engine
│   │   └── agent/          # AI evaluation agent
│   ├── run/                 # Run lifecycle management
│   └── tools/               # Shared utilities
└── .github/
    └── workflows/           # CI/CD pipelines
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the GNU GPL v3.0. See [LICENSE](LICENSE) for the full license text.

## Acknowledgments

- Built with [urfave/cli](https://github.com/urfave/cli) for CLI interface
- Uses [langchaingo](https://github.com/tmc/langchaingo) for AI evaluation
- Database management with [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)

## Contact

For questions, issues, or suggestions, please open an issue on GitHub.

---
