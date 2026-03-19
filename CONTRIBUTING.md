# Contributing

Thank you for contributing to CochaBench.

## Before You Start

- Open an issue for bugs, regressions, or larger feature ideas before starting substantial work.
- Keep changes focused. Small, reviewable pull requests are preferred over broad refactors.
- Do not include unrelated formatting or cleanup changes in feature or bugfix pull requests.

## Development Setup

Prerequisites:

- Go 1.25.5 or higher
- Git

Clone the repository and run the test suite:

```bash
git clone https://github.com/EinfachNiklas/cochabench.git
cd cochabench
go test ./...
```

## Project Expectations

- Preserve the existing CLI behavior unless the change explicitly updates the contract.
- Add or update tests when changing behavior.
- Keep documentation aligned with the implemented commands and flags.
- Avoid committing secrets, tokens, local config files, or generated artifacts.

## Pull Requests

Before opening a pull request:

- Run `go test ./...`
- Update documentation if user-facing behavior changed
- Confirm that no generated binaries or local environment files are included

Pull requests should include:

- A clear description of the problem
- A concise summary of the change
- Notes about behavior changes, migration concerns, or follow-up work

## Code Style

- Follow standard Go formatting conventions
- Prefer clear, explicit code over clever abstractions
- Keep dependencies minimal and justified

## Questions

If you are unsure whether a change fits the project, open an issue first.
