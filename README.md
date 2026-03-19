# CochaBench

![GitHub Actions](https://github.com/EinfachNiklas/cochabench/workflows/Run%20Tests/badge.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/EinfachNiklas/cochabench)

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

### 5. Start Working

```bash
cochabench run start --id <run-id>
```

Work on your solution in the `solutions/<run-id>/` directory.

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
- **Quality Score**: AI-evaluated code quality (0-100)
- **Maintainability Score**: AI-evaluated code maintainability (0-100)
- **Security Score**: AI-evaluated code security (0-100)

### AI support

Currently, only Anthropic Models (Claude) are supported.

## Environment Variables

CochaBench supports the following environment variables:

- `GITHUB_TOKEN`: GitHub personal access token for API requests (optional, required to connect to private challenge server repos)
- `LLM_API_KEY`: LLM API KEY for ai evaluation

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

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source. Please check the repository for license details.

## Acknowledgments

- Built with [urfave/cli](https://github.com/urfave/cli) for CLI interface
- Uses [langchaingo](https://github.com/tmc/langchaingo) for AI evaluation
- Database management with [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)

## Contact

For questions, issues, or suggestions, please open an issue on GitHub.

---
