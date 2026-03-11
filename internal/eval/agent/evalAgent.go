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
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

const (
	EvaluationRuns = 3
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
		if c.LLM_BASE_PATH != "" {
			opts = append(opts, anthropic.WithBaseURL(c.LLM_BASE_PATH))
		}
		llm, err = anthropic.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("Failed to setup Anthropic %s: %v\n", c.LLM_MODEL, err)
		}
	default:
		return nil, fmt.Errorf("LLM Provider %s is not supported", c.LLM_PROVIDER)
	}
	return &llm, nil
}

func getEvalAgent() (*agents.OneShotZeroAgent, error) {
	llm, err := getLLM()
	if err != nil {
		return nil, fmt.Errorf("Failed to set up EvalAgent: %v\n", err)
	}

	agentTools := []tools.Tool{
		DirectoryLister{},
		FileReader{},
	}

	customPrompt := prompts.PromptTemplate{
		Template: `
			You are an experienced Code Reviewer.
			
			Goal: Evaluate a coding Project regarding software quality (Readability, Structure, Adherence to Coding Conventions), maintainability (Maintainability, Modularity, Extendability) and security (Security Aspects, potential weaknesses) and generate a score from 1-10 for each category.

			The test files were already provided ans must not influence your score decision.

			The coding Project is located in the directory {{.tmp_dir}}

			You have access to the following tools:
			{{.tool_descriptions}}
			
			IMPORTANT: You MUST use the following format for your response:

			Thought: [your reasoning about what to do next]
			Action: [the action to take, must be one of: {{.tool_names}}]
			Action Input: [the input to the action]
			Observation: [the result of the action]
			... (this Thought/Action/Action Input/Observation can repeat N times)
			Thought: I now know the final answer
			Final Answer: [The scores in json]

			Your final answer must adhere to the following json template with your scores:

			{
				"quality": YOUR_SCORE,
				"maintainability": YOUR_SCORE,
				"security": YOUR_SCORE
			}

			Do NOT use XML tags like <function_calls> or <invoke>. Use the exact format shown above.

			Begin!

			{{.agent_scratchpad}}
		`,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: []string{"tmp_dir", "agent_scratchpad"},
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

func runSingleEvaluation(ctx context.Context, cScores chan []float64, cErr chan error, run int, tmpDir string) {
	agent, err := getEvalAgent()
	if err != nil {
		cErr <- fmt.Errorf("Error on run %d: %v\n", run, err)
		return
	}
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(20))

	responses, err := executor.Call(ctx, map[string]any{"tmp_dir": tmpDir}, chains.WithTemperature(0.0))
	if err != nil {
		cErr <- fmt.Errorf("Error on run %d: %v\n", run, err)
		return
	}

	output, ok := responses["output"].(string)
	if !ok {
		cErr <- fmt.Errorf("Failed to extract output from agent response on run %d", run)
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
		cErr <- fmt.Errorf("Failed to parse JSON output on run %d: %w\nOutput was: %s", run, err, output)
		return
	}

	cScores <- []float64{runResult.Quality, runResult.Maintainability, runResult.Security}
}

func (evaluator *Evaluator) Evaluate(tmpDir string) (*EvaluatorResult, error) {
	fmt.Printf("Starting AI Evaluation (%d runs)...\n", EvaluationRuns)

	var qualitySum, maintainabilitySum, securitySum float64

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cScores := make(chan []float64, EvaluationRuns)
	cErrors := make(chan error, EvaluationRuns)
	for run := 1; run <= EvaluationRuns; run++ {
		fmt.Printf("Run %d/%d...\n", run, EvaluationRuns)
		go runSingleEvaluation(ctx, cScores, cErrors, run, tmpDir)
	}

	for range EvaluationRuns {
		select {
		case runScores := <-cScores:
			qualitySum += runScores[0]
			maintainabilitySum += runScores[1]
			securitySum += runScores[2]
		case err := <-cErrors:
			cancel()
			return nil, fmt.Errorf("Error when running Evaluator: %v\n", err)
		}
	}
	avgResult := &EvaluatorResult{
		Quality:         qualitySum / float64(EvaluationRuns),
		Maintainability: maintainabilitySum / float64(EvaluationRuns),
		Security:        securitySum / float64(EvaluationRuns),
	}

	fmt.Printf("AI Evaluation Result: Quality=%.2f, Maintainability=%.2f, Security=%.2f\n",
		avgResult.Quality, avgResult.Maintainability, avgResult.Security)
	fmt.Println("Finished AI Evaluation")
	return avgResult, nil

}
