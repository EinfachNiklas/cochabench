# CochaBench

> A comprehensive coding challenge benchmark suite for developers and AI coding agents

[![GitHub Actions](https://github.com/EinfachNiklas/cochabench/workflows/Run%20Tests/badge.svg)](https://github.com/EinfachNiklas/cochabench/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/EinfachNiklas/cochabench)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/EinfachNiklas/cochabench/.svg)](https://pkg.go.dev/github.com/EinfachNiklas/cochabench/)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

---

## What is CochaBench?

**CochaBench** is a CLI tool designed to evaluate and compare the performance of developers and AI coding agents across standardized coding challenges. It supports multiple programming languages and provides both automated test execution and AI-powered code quality assessment.

---

## Key Features

- **Multi-Language Support** - Challenges available in JavaScript, Python, and Go
- **Standardized Evaluation** - Consistent metrics across all challenges
- **AI-Powered Assessment** - Code quality, maintainability, and security scoring
- **Flexible Workflow** - Initialize, start, stop, and evaluate coding attempts
- **Persistent Storage** - SQLite-based tracking of all runs and evaluations

---

## How It Works

```
1. Download Challenge    →    cochabench challenge get <id>
2. Initialize Run        →    cochabench run init --name "my-attempt"
3. Start Timer           →    cochabench run start --id <run-id>
4. Write Your Solution   →    Edit files in solutions/<run-id>/
5. Stop Timer            →    cochabench run stop --id <run-id>
6. Evaluate              →    cochabench run eval --runID <run-id>
```

---

## Quick Links

- [Installation](installation.md) - Get CochaBench up and running
- [Quick Start](quickstart.md) - Complete your first challenge
- [CLI Reference](cli-reference.md) - Full command documentation
- [Configuration](configuration.md) - Customize your setup
- [Data Management](data-management.md) - Merge and analyze runs across challenges
- [Evaluation](evaluation.md) - Understanding metrics and AI assessment

---

## Evaluation Metrics

| Metric | Description | Range |
|--------|-------------|-------|
| Test Results | Pass/fail counts from test execution | Count |
| Execution Time | Duration of test run | Time |
| Quality Score | Code readability and structure | 1-10 |
| Maintainability | Modularity and extendability | 1-10 |
| Security Score | Input validation and vulnerability assessment | 1-10 |

---

## Supported AI Providers

CochaBench integrates with major AI providers for code evaluation:

- **Anthropic** (Claude)
- **OpenAI** (GPT, Codex)
- **Google** (Gemini)

---

## License

This project is licensed under the [GNU GPL v3.0](https://github.com/EinfachNiklas/cochabench/blob/main/LICENSE).
