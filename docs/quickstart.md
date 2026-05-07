# Quick Start

This guide walks you through completing your first coding challenge with CochaBench.

---

## Step 1: Initialize Configuration

First, create the CochaBench configuration file:

```bash
cochabench config init
```

This creates `~/.config/cochabench/config.json` with default settings.

---

## Step 2: Browse Available Challenges

List all available coding challenges:

```bash
cochabench challenge list
```

Output example:
```
Release: v1.2.0

ID                     | Title                               | Language   | Difficulty
-----------------------+-------------------------------------+------------+-----------
binary-protocol-parser | Binary Protocol Parser              | go         | medium    
graph-pathfinding      | Graph Pathfinding mit Constraints   | javascript | medium    
lru-cache-with-ttl     | LRU Cache with TTL                  | python     | medium    
markdown-renderer      | Markdown Renderer mit Plugin-System | javascript | medium    
task-scheduler         | Task Scheduler                      | go         | medium    
web-crawler            | Web Crawler                         | go         | medium
```

---

## Step 3: Download a Challenge

Download a specific challenge by ID:

```bash
cochabench challenge get graph-pathfinding
```

Or download all available challenges:

```bash
cochabench challenge get all
```

A Challenge Directory is created for each downloaded challenge.

---

## Step 4: Navigate to Challenge Directory

```bash
cd graph-pathfinding
```

Each challenge directory contains:
```
graph-pathfinding /
├── src/                    # Starter code template
├── test/                   # Pre-written test suite
└── challenge.config.json   # Challenge metadata
```

---

## Step 5: Initialize a Run

Create a new run to track your attempt:

```bash
cochabench run init --name "my-first-attempt"
```

Output:
```
Initialized run my-first-attempt[abc123-...] successfully
```

This creates:
- A `solutions/<run-id>/` directory with a copy of the starter code
- An entry in the local `cochabench.db` SQLite database

---

## Step 6: Start the Run

Begin your timed attempt:

```bash
cochabench run start --id <run-id>
```

The run is now stared!

---

## Step 7: Write Your Solution

Edit the files in `solutions/<run-id>/`:

Implement your solution. 

>[!NOTE]
>The tests are in a separate directory (`test/`) and can not be run directly. Do not modify these tests. They will be used for the evaluation.

---

## Step 8: Stop the Run

When you're done, stop the run:

```bash
cochabench run stop --id <run-id>
```

This records your completion time.

---

## Step 9: Evaluate Your Solution

Run the evaluation:

```bash
cochabench run eval --runID <run-id>
```

### Evaluation Output

```
RunID                                | RunName  | Status | StartTime           | EndTime             | Duration     | TimedOut | Total | Passed | Failed | Quality | Maintainability | Security
-------------------------------------+----------+--------+---------------------+---------------------+--------------+----------+-------+--------+--------+---------+-----------------+---------
24630941-d437-457a-972a-f8536061da6b | STATIC_3 | F      | 2026-04-09 11:31:56 | 2026-04-09 11:33:35 | 159.3045933s | false    | 22    | 22     | 0      | 7.67    | 7.33            | 5.67
```

---

## Step 10: Review Results

### Check All Runs

List all your runs for the current challenge:

```bash
cochabench run list
```

### Understanding Scores

| Score | Meaning |
|-------|---------|
| 1-3 | Poor - Significant issues |
| 4-6 | Fair - Room for improvement |
| 7-9 | Good - Solid implementation |
| 10 | Excellent - Exemplary code |

---

## Quick Reference

```bash
# Full workflow in one view
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

---

## Tips

- **Use `--no-ai-eval`** for faster evaluation without AI scoring
- **Cancel a run** with `cochabench run cancel --id <run-id>` if needed

---

## Next Steps

- [CLI Reference](cli-reference.md) - Full command documentation
- [Configuration](configuration.md) - Customize AI provider and settings
- [Data Management](data-management.md) - Merge and analyze runs across challenges
- [Evaluation](evaluation.md) - Deep dive into scoring metrics
