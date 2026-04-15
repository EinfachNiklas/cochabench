# Configuration

CochaBench uses a JSON configuration file and environment variables.

---

## Configuration File

**Location**: `~/.config/cochabench/config.json`

### Initialize Configuration

```bash
cochabench config init
```

### Default Configuration

```json
{
  "LLM_PROVIDER": "anthropic",
  "LLM_BASE_PATH": "https://api.anthropic.com/v1",
  "LLM_MODEL": "claude-sonnet-4-6",
  "CHALLENGE_SERVER": "https://api.github.com/repos/EinfachNiklas/cochabench-challenges/"
}
```

---

## Configuration Keys

### LLM_PROVIDER

The AI provider for code evaluation.

| Value | Provider |
|-------|----------|
| `anthropic` | Anthropic (Claude) |
| `openai` | OpenAI (GPT, Codex) |
| `google` | Google (Gemini) |


### LLM_BASE_PATH

Base URL for the AI provider API. Used for custom endpoints or proxies.

| Provider | Default |
|----------|---------|
| Anthropic | `https://api.anthropic.com/v1` |
| OpenAI | `https://api.openai.com/v1` |
| Google | *(uses default)* |


### LLM_MODEL

The specific model to use for AI evaluation (excerpt):

**Anthropic Models**:
- `claude-sonnet-4-6` (default)
- `claude-opus-4-6`
- `claude-haiku-4-5-20251001`

**OpenAI Models**:
- `gpt-4`
- `gpt-4-turbo`
- `gpt-3.5-turbo`

**Google Models**:
- `gemini-pro`
- `gemini-1.5-pro`


### CHALLENGE_SERVER

URL of the challenge repository API. Default points to the official CochaBench challenges.

Url Template for GitHub: `https://api.github.com/repos/your-org/your-repo/`

Url Template for GitHub Enterprise: `https://your-enterprise-domain/api/v3/repos/your-org/your-repo/`

---

## Environment Variables

Environment variables are used for sensitive values that should not be stored in config files.

### LLM_API_KEY

**Required** for AI evaluation.

API key from your AI provider:

```bash
# Linux/macOS
export LLM_API_KEY="your-api-key"

# Windows (PowerShell)
$env:LLM_API_KEY="your-api-key"

# Windows (CMD)
set LLM_API_KEY=your-api-key
```

### GITHUB_TOKEN

**Optional**. Required only for private challenge repositories.

>[!IMPORTANT]
>GitHub imposes rate limits on requests to public repos. Creating a Personal Access Token for GitHub (even for the default repository) increases these limits drastically and prevents problems during the usage of Cochabench.

```bash
# Linux/macOS
export GITHUB_TOKEN="your-gh-token"

# Windows (PowerShell)
$env:GITHUB_TOKEN="your-gh-token"

# Windows (CMD)
set GITHUB_TOKEN=your-gh-token
```

---

## Managing Configuration

### View All Settings

```bash
cochabench config show
```
See [Reference](cli-reference.md#config-show)

### Get Single Value

```bash
cochabench config get LLM_MODEL
```
See [Reference](cli-reference.md#config-get)

### Set Value

```bash
cochabench config set LLM_MODEL gpt-4
```
See [Reference](cli-reference.md#config-set)

---

## Provider Setup Examples

### Anthropic (Claude)

```bash
cochabench config set LLM_PROVIDER anthropic
cochabench config set LLM_MODEL claude-sonnet-4-6
export LLM_API_KEY="your-api-key"
```

### OpenAI

```bash
cochabench config set LLM_PROVIDER openai
cochabench config set LLM_MODEL gpt-4
export LLM_API_KEY="your-api-key"
```

### Google (Gemini)

```bash
cochabench config set LLM_PROVIDER google
cochabench config set LLM_MODEL gemini-pro
export LLM_API_KEY="your-api-key"
```

---

## Troubleshooting

### "Could not initialize Anthropic model"

- Verify `LLM_API_KEY` is set
- Check the API key is valid
- Ensure `LLM_MODEL` is a valid model name

### "Could not fetch challenge list"

- Check `CHALLENGE_SERVER` URL is correct
- For private repos, verify `GITHUB_TOKEN` is set and has appropriate permissions

### Rate Limit Errors

- See [GITHUB_TOKEN](#github_token)
