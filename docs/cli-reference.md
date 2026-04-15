# CLI Reference

Complete reference for all CochaBench commands.

## Global Options

```bash
cochabench --version    # Show version
cochabench --help       # Show help
```

---

## Challenge Commands

Manage coding challenges.

### `challenge list`

List all available challenges from the challenge server.

```bash
cochabench challenge list
cochabench c l              # Short form
```

**Output**: Table with challenge ID, title, language, and difficulty.

---

### `challenge get`

Download a specific challenge.

```bash
cochabench challenge get <challenge-id>
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `challenge-id` | ID of the challenge to download |

**Example**:
```bash
cochabench challenge get graph-pathfinding 
```

---

### `challenge get all`

Download all available challenges.

```bash
cochabench challenge get all
cochabench c get a          # Short form
```

---

## Run Commands

Manage challenge runs and attempts.

### `run init`

Initialize a new run for the current challenge.

```bash
cochabench run init [options]
cochabench r init           # Short form
```

**Flags**:

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `--name` | `-n` | UUID | Name for the run |
| `--print-id-only` | `--id-only` | false | Only print the run ID (for automation) |

**Example**:
```bash
cochabench run init --name "attempt-1"
cochabench run init --id-only    # Outputs just the UUID
```

**Behavior**:
- Creates `solutions/<run-id>/` directory
- Copies starter code from `src/` to the solution directory
- Records the run in `cochabench.db`

---

### `run start`

Start a run.

```bash
cochabench run start --id <run-id>
```

**Flags**:

| Flag | Alias | Description |
|------|-------|-------------|
| `--id` | `-i` | Run ID to start |

**Requirements**:
- Run must be in "Initialized" (I) or "Canceled" (C) status

---

### `run stop`

Stop a run.

```bash
cochabench run stop --id <run-id>
```

**Flags**:

| Flag | Alias | Description |
|------|-------|-------------|
| `--id` | `-i` | Run ID to stop |

**Requirements**:
- Run must be in "Running" (R) status

**Behavior**:
- Records end time and calculates duration
- Sets status to "Finished" (F)

---

### `run cancel`

Cancel a running attempt without finishing.

```bash
cochabench run cancel --id <run-id>
```

**Flags**:

| Flag | Alias | Description |
|------|-------|-------------|
| `--id` | `-i` | Run ID to cancel |

**Requirements**:
- Run must be in "Running" (R) status

**Behavior**:
- Sets status to "Canceled" (C)
- Can be restarted with `run start`

---

### `run list`

List all runs for the current challenge.

```bash
cochabench run list
```

**Output**: Table with run details including status, times, and evaluation scores.

---

### `run eval`

Evaluate a completed run.

```bash
cochabench run eval --runID <run-id> [options]
```

**Flags**:

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `--runID` | `-i` | - | Run ID to evaluate (required) |
| `--path` | `-p` | `./` | Path to challenge directory |
| `--debug` | `-d` | false | Keep temporary files after evaluation |
| `--no-ai-eval` | `--no-ai` | false | Skip AI evaluation |
| `--number-of-agents` | `-n` | 3 | Number of AI evaluation runs |
| `--timeout` | `-t` | 5m | Test execution timeout |

**Requirements**:
- Run must be in "Finished" (F) status

**Examples**:
```bash
# Standard evaluation
cochabench run eval --runID abc123

# Fast evaluation without AI
cochabench run eval --runID abc123 --no-ai-eval

# Debug mode with custom timeout
cochabench run eval --runID abc123 --debug --timeout 10m

# Single AI evaluation run
cochabench run eval --runID abc123 -n 1
```

---

## Config Commands

Manage CochaBench configuration.

### `config init`

Initialize the configuration file.

```bash
cochabench config init
```

**Behavior**:
- Creates `~/.config/cochabench/config.json` if it doesn't exist
- Uses default values for all settings

---

### `config show`

Display all configuration values.

```bash
cochabench config show
```

**Output**: 
```
Key              | Value
-----------------+------------------------------------------------------------------
CHALLENGE_SERVER | https://api.github.com/repos/EinfachNiklas/cochabench-challenges/
LLM_BASE_PATH    | https://api.anthropic.com/v1
LLM_MODEL        | anthropic--claude-4.6-sonnet
LLM_PROVIDER     | anthropic
```

---

### `config get`

Get a specific configuration value.

```bash
cochabench config get <key>
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `key` | Configuration key to retrieve |

**Example**:
```bash
cochabench config get LLM_MODEL
```

---

### `config set`

Set a configuration value.

```bash
cochabench config set <key> <value>
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `key` | Configuration key to set |
| `value` | New value |

**Example**:
```bash
cochabench config set LLM_PROVIDER openai
cochabench config set LLM_MODEL gpt-4
```

---

## Run Status Codes

| Code | Status | Description |
|------|--------|-------------|
| I | Initialized | Run created, not started |
| R | Running | Timer active |
| C | Canceled | Run was canceled |
| F | Finished | Run completed, ready for evaluation |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LLM_API_KEY` | API key for AI evaluation |
| `GITHUB_TOKEN` | GitHub token for private challenge repos |
