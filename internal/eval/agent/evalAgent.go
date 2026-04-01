package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cfg "github.com/EinfachNiklas/cochabench/internal/config"
	projectTools "github.com/EinfachNiklas/cochabench/internal/tools"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

type Evaluator struct{}

type EvaluatorResult struct {
	Quality         float64 `json:"quality"`
	Maintainability float64 `json:"maintainability"`
	Security        float64 `json:"security"`
}

func csvFromTools(tools []tools.Tool, mode int) string {
	var values []string
	for _, v := range tools {
		if mode == 1 {
			values = append(values, v.Name())
		} else if mode == 2 {
			values = append(values, fmt.Sprintf("%s: %s", v.Name(), v.Description()))
		} else {
			return ""
		}
	}
	return strings.Join(values, "; ")
}

func getLLM() (*llms.Model, error) {
	env, err := projectTools.LoadEnv()
	if err != nil {
		return nil, err
	}

	c, err := cfg.GetConfig()
	if err != nil {
		return nil, err
	}

	var llm llms.Model
	switch c.LLM_PROVIDER {
	case "anthropic":
		opts := []anthropic.Option{
			anthropic.WithToken(env.LLM_API_KEY),
			anthropic.WithModel(c.LLM_MODEL),
		}
		if len(c.LLM_BASE_PATH) != 0 {
			opts = append(opts, anthropic.WithBaseURL(c.LLM_BASE_PATH))
		}
		llm, err = anthropic.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("Could not initialize Anthropic model %s: %w", c.LLM_MODEL, err)
		}
	case "openai":
		opts := []openai.Option{
			openai.WithToken(env.LLM_API_KEY),
			openai.WithModel(c.LLM_MODEL),
		}
		if len(c.LLM_BASE_PATH) != 0 {
			opts = append(opts, openai.WithBaseURL(c.LLM_BASE_PATH))
		}
		llm, err = openai.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("Could not initialize OpenAI model %s: %w", c.LLM_MODEL, err)
		}
	case "google":
		opts := []googleai.Option{
			googleai.WithAPIKey(env.LLM_API_KEY),
			googleai.WithDefaultModel(c.LLM_MODEL),
		}
		if len(c.LLM_BASE_PATH) != 0 {
			opts = append(opts, googleai.WithCloudLocation(c.LLM_BASE_PATH))
		}
		llm, err = googleai.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("Could not initialize Google model %s: %w", c.LLM_MODEL, err)
		}
	default:
		return nil, fmt.Errorf("Unsupported LLM provider: %s", c.LLM_PROVIDER)
	}
	return &llm, nil
}

func getEvalAgent() (*agents.OneShotZeroAgent, error) {
	llm, err := getLLM()
	if err != nil {
		return nil, fmt.Errorf("Could not set up AI evaluator: %w", err)
	}

	agentTools := []tools.Tool{
		DirectoryLister{},
		FileReader{},
	}

	customPrompt := prompts.PromptTemplate{
		Template: `You are a senior code reviewer. Your evaluation determines deployment decisions.

Evaluate the code in {{.tmp_dir}} on three categories, each scored 1-10.

## Scoring Rubric

### Quality (Readability, Structure, Conventions)
- 1-3: Inconsistent naming, no structure, ignores language conventions
- 4-6: Mostly readable, some structural issues, partially follows conventions
- 7-9: Clean naming, clear structure, follows idiomatic patterns consistently
- 10: Exemplary — could serve as a reference implementation

### Maintainability (Modularity, Extendability, Clarity of Intent)
- 1-3: Monolithic, tightly coupled, no separation of concerns
- 4-6: Some modularity, but extending functionality would require significant refactoring
- 7-9: Well-separated concerns, clear interfaces, easy to extend
- 10: Highly modular with clear extension points and minimal coupling

### Security (Input Validation, Data Handling, Known Vulnerability Patterns)
- 1-3: Unvalidated inputs, hardcoded secrets, injection vulnerabilities
- 4-6: Basic validation present, but gaps in edge cases or error handling
- 7-9: Consistent input validation, proper error handling, no obvious vulnerabilities
- 10: Defense in depth — validates at every boundary, handles all edge cases

## Rules
- IGNORE all test files (*_test.go, *_test.py, *.test.js, *.test.ts, *.spec.*). They are pre-provided and NOT part of the evaluation.
- Score ONLY the implementation code written by the developer.
- Before assigning each score, you MUST state at least one concrete positive finding and one concrete issue (or explicitly state "no issues found").
- Scores must be integers from 1 to 10.

{{if .diff}}
## Developer Changes (Diff)
The unified diff below shows what the developer changed from the original template.
Focus your evaluation PRIMARILY on this changed/new code.
Only use tools for additional context when needed (e.g., understanding how a changed function integrates).

{{.diff}}

IMPORTANT: Score the CHANGES, not the pre-existing template code.
{{else}}
## Exploration
No diff available. Use directory_lister and file_reader to explore and review the project code.
{{end}}

## Tools
{{.tool_descriptions}}

## Response Format
You MUST use this exact format:

Thought: [your reasoning]
Action: [must be one of: {{.tool_names}}]
Action Input: [input for the action]
Observation: [result]
... (repeat as needed)
Thought: I now know the final answer
Final Answer: [JSON]

Your Final Answer MUST be this JSON (scores as integers, reasoning is mandatory):
{
  "quality": YOUR_SCORE,
  "quality_reasoning": "one sentence justification",
  "maintainability": YOUR_SCORE,
  "maintainability_reasoning": "one sentence justification",
  "security": YOUR_SCORE,
  "security_reasoning": "one sentence justification"
}

Do NOT use XML tags. Use the exact format above.

Begin!

{{.agent_scratchpad}}
`,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: []string{"tmp_dir", "diff", "agent_scratchpad"},
		PartialVariables: map[string]any{
			"tool_names":        csvFromTools(agentTools, 1),
			"tool_descriptions": csvFromTools(agentTools, 2),
		},
	}

	// Custom error handler for parsing errors
	parserErrorHandler := agents.NewParserErrorHandler(func(err string) string {
		return fmt.Sprintf("Invalid output format. You used the wrong format. Please follow the EXACT format:\nThought: [your reasoning]\nAction: [tool name]\nAction Input: [input for tool]\n\nDo NOT use XML tags. Error was: %s", err)
	})

	agent := agents.NewOneShotAgent(
		*llm,
		agentTools,
		agents.WithPrompt(customPrompt),
		agents.WithParserErrorHandler(parserErrorHandler),
	)
	return agent, nil
}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func runSingleEvaluation(ctx context.Context, cScores chan []float64, cErr chan error, run int, tmpDir string, diff string) {
	agent, err := getEvalAgent()
	if err != nil {
		cErr <- fmt.Errorf("AI evaluation run %d failed: %w", run, err)
		return
	}
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(20))

	responses, err := executor.Call(ctx, map[string]any{"tmp_dir": tmpDir, "diff": diff}, chains.WithTemperature(0.0))
	if err != nil {
		cErr <- fmt.Errorf("AI evaluation run %d failed: %w", run, err)
		return
	}

	output, ok := responses["output"].(string)
	if !ok {
		cErr <- fmt.Errorf("Could not extract AI evaluation output from run %d", run)
		return
	}

	// Remove markdown code blocks if present (```json ... ``` or ``` ... ```)
	output = strings.TrimSpace(output)
	if strings.HasPrefix(output, "```") {
		output = strings.TrimPrefix(output, "```json")
		output = strings.TrimPrefix(output, "```")
		output = strings.TrimSpace(output)
		if idx := strings.LastIndex(output, "```"); idx != -1 {
			output = output[:idx]
		}
		output = strings.TrimSpace(output)
	}

	var runResult EvaluatorResult
	if err := json.Unmarshal([]byte(output), &runResult); err != nil {
		cErr <- fmt.Errorf("Could not parse AI evaluation output for run %d: %w", run, err)
		return
	}

	cScores <- []float64{runResult.Quality, runResult.Maintainability, runResult.Security}
}

func (evaluator *Evaluator) Evaluate(tmpDir string, diff string, numEvalRuns int) (*EvaluatorResult, error) {
	fmt.Printf("Starting AI Evaluation (%d runs)...\n", numEvalRuns)

	var qualitySum, maintainabilitySum, securitySum float64

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cScores := make(chan []float64, numEvalRuns)
	cErrors := make(chan error, numEvalRuns)
	for run := 1; run <= numEvalRuns; run++ {
		fmt.Printf("Run %d/%d...\n", run, numEvalRuns)
		go runSingleEvaluation(ctx, cScores, cErrors, run, tmpDir, diff)
	}

	for range numEvalRuns {
		select {
		case runScores := <-cScores:
			qualitySum += runScores[0]
			maintainabilitySum += runScores[1]
			securitySum += runScores[2]
		case err := <-cErrors:
			cancel()
			return nil, fmt.Errorf("Could not complete AI evaluation: %w", err)
		}
	}
	avgResult := &EvaluatorResult{
		Quality:         qualitySum / float64(numEvalRuns),
		Maintainability: maintainabilitySum / float64(numEvalRuns),
		Security:        securitySum / float64(numEvalRuns),
	}

	fmt.Printf("AI Evaluation Result: Quality=%.2f, Maintainability=%.2f, Security=%.2f\n",
		avgResult.Quality, avgResult.Maintainability, avgResult.Security)
	fmt.Println("Finished AI Evaluation")
	return avgResult, nil

}
