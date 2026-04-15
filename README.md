# CochaBench

![GitHub Actions](https://github.com/EinfachNiklas/cochabench/workflows/Run%20Tests/badge.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/EinfachNiklas/cochabench)
[![Go Reference](https://pkg.go.dev/badge/github.com/EinfachNiklas/cochabench/.svg)](https://pkg.go.dev/github.com/EinfachNiklas/cochabench/)

**CochaBench** is a comprehensive coding challenge benchmark suite designed to evaluate and compare the performance of developers and AI coding agents across multiple programming languages.

## Features

- **Multi-Language Support**: Challenges available in JavaScript, Python, and Go
- **Standardized Evaluation**: Test execution with pass/fail rates and execution time tracking
- **AI-Powered Assessment**: Code quality, maintainability, and security scoring (1-10)
- **Flexible Workflow**: Initialize, start, stop, and evaluate coding challenge attempts
- **Persistent Storage**: SQLite-based tracking of all runs and evaluations

## Installation

```bash
go install github.com/EinfachNiklas/cochabench/cmd/cochabench@latest
```

**Prerequisites**: Go 1.25.5+, Git, and language-specific tools (npm/python/go) for challenge evaluation.

## Quick Start

```bash
cochabench config init                    # One-time setup
cochabench challenge list                 # Browse challenges
cochabench challenge get <id>             # Download challenge
cd <challenge-dir>                        # Enter challenge
cochabench run init --name "attempt"      # Create run
cochabench run start --id <run-id>        # Start run
# ... write your solution ...
cochabench run stop --id <run-id>         # Stop run
cochabench run eval --runID <run-id>      # Evaluate
```

## Documentation

Full documentation is available on [GitHub Pages](https://einfachniklas.github.io/cochabench).

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## Security

CochaBench executes code locally on your machine. See [SECURITY.md](SECURITY.md) for details and vulnerability reporting.

## License

This project is licensed under the GNU GPL v3.0. See [LICENSE](LICENSE) for the full license text.

## Acknowledgments

- Built with [urfave/cli](https://github.com/urfave/cli) for CLI interface
- Uses [langchaingo](https://github.com/tmc/langchaingo) for AI evaluation
- Database management with [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)
