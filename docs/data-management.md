# Data Management

CochaBench allows you to merge run data from multiple challenges into a single database for centralized analysis.

---
## Merged Database Overview

The merged database (`cochabenchMerged.db`) combines:
- **Challenges**: List of all challenge IDs from merged directories.
- **Runs**: All runs from all challenges, with foreign keys linking to their respective challenges.

### Database Schema

#### `challenges` Table
| Column | Type | Description |
|--------|------|-------------|
| `challengeId` | CHAR(256) | Unique identifier for the challenge (primary key) |

#### `runs` Table
| Column | Type | Description |
|--------|------|-------------|
| `runId` | CHAR(36) | Unique run identifier (primary key) |
| `challengeId` | CHAR(256) | Foreign key to `challenges.challengeId` |
| `runName` | VARCHAR(256) | Human-readable name of the run |
| `runStatus` | CHAR(1) | Status code (I/R/C/F) |
| `startTime` | TIMESTAMP | When the run was started |
| `endTime` | TIMESTAMP | When the run was stopped |
| `duration` | INTEGER | Duration in nanoseconds |
| `testTimedOut` | BOOLEAN | Whether the run timed out |
| `numTotalTests` | INTEGER | Total tests executed |
| `numPassedTests` | INTEGER | Tests passed |
| `numFailedTests` | INTEGER | Tests failed |
| `qualityScore` | DECIMAL(15,2) | AI-evaluated quality score (1-10) |
| `maintainabilityScore` | DECIMAL(15,2) | AI-evaluated maintainability score (1-10) |
| `securityScore` | DECIMAL(15,2) | AI-evaluated security score (1-10) |

---
## Use Cases

### Aggregate Results Across Challenges
Merge all your challenge attempts to analyze patterns:
```bash
cochabench data merge --path ~/my-challenges
```

---
## How It Works

1. **Validation**: The command checks that the provided path does **not** contain a `challenge.config.json` (i.e., it must be a parent directory of challenges, not a challenge itself).
2. **Discovery**: Scans the directory for subdirectories with `challenge.config.json` files.
3. **Aggregation**: For each challenge:
   - Reads the `challengeId` from `challenge.config.json`.
   - Opens the challenge's `cochabench.db` and extracts all runs.
   - Inserts the challenge into the `challenges` table (ignores duplicates).
   - Inserts runs into the `runs` table (updates existing runs with the same `runId`).
4. **Output**: Creates `cochabenchMerged.db` in the specified directory.

---