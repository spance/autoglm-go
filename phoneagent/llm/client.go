package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spance/autoglm-go/phoneagent/definitions"
	"github.com/spance/autoglm-go/phoneagent/helper"
)

type ModelClient struct {
	config *definitions.ModelConfig
	client *openai.Client
}

func NewModelClient(cfg *definitions.ModelConfig) *ModelClient {
	if cfg == nil {
		cfg = &definitions.ModelConfig{}
	}
	openaiCfg := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		openaiCfg.BaseURL = cfg.BaseURL
	}

	return &ModelClient{
		config: cfg,
		client: openai.NewClientWithConfig(openaiCfg),
	}
}

type ModelResponse struct {
	Thinking          string
	Action            string
	ToolCalls         []openai.ToolCall
	RawContent        string
	TimeToFirstToken  *float64
	TimeToThinkingEnd *float64
	TotalTime         float64
}

func (c *ModelClient) Request(ctx context.Context, messages []openai.ChatCompletionMessage) (*ModelResponse, error) {
	startTime := time.Now()

	var (
		timeToFirstToken  *float64
		timeToThinkingEnd *float64
	)

	req := openai.ChatCompletionRequest{
		Model:               c.config.ModelName,
		Messages:            messages,
		MaxCompletionTokens: c.config.MaxTokens,
		Temperature:         c.config.Temperature,
		TopP:                c.config.TopP,
		FrequencyPenalty:    c.config.FrequencyPenalty,
		Tools:               definitions.GetPhoneAgentTools(),
		ToolChoice:          "auto",
		Stream:              false,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("CreateChatCompletion error")
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	choice := resp.Choices[0]
	totalTime := time.Since(startTime).Seconds()

	// Record timing
	t := totalTime
	timeToFirstToken = &t
	timeToThinkingEnd = &t

	// Extract thinking from content
	thinking := strings.TrimSpace(choice.Message.Content)
	// if thinking != "" {
	// 	log.Info().Msg(thinking)
	// }

	// Extract tool calls
	var toolCalls []openai.ToolCall
	var action string

	if len(choice.Message.ToolCalls) > 0 {
		toolCalls = choice.Message.ToolCalls
		// Format action string from tool call for logging
		firstCall := choice.Message.ToolCalls[0]
		action = fmt.Sprintf("%s(%s)", firstCall.Function.Name, firstCall.Function.Arguments)
	}

	printMetrics(
		c.config.Lang,
		timeToFirstToken,
		timeToThinkingEnd,
		totalTime,
	)

	return &ModelResponse{
		Thinking:          thinking,
		Action:            action,
		ToolCalls:         toolCalls,
		RawContent:        choice.Message.Content,
		TimeToFirstToken:  timeToFirstToken,
		TimeToThinkingEnd: timeToThinkingEnd,
		TotalTime:         totalTime,
	}, nil
}

func printMetrics(lang string, firstToken *float64, thinkingEnd *float64, total float64) {
	log.Info().Msg("")
	log.Info().Msg(strings.Repeat("=", 50))
	log.Info().Msg("⏱️  " + helper.GetMessage("performance_metrics", lang))
	log.Info().Msg(strings.Repeat("-", 50))

	if firstToken != nil {
		log.Info().Msgf("%s: %.3fs", helper.GetMessage("time_to_first_token", lang), *firstToken)
	}
	if thinkingEnd != nil {
		log.Info().Msgf("%s: %.3fs", helper.GetMessage("time_to_thinking_end", lang), *thinkingEnd)
	}
	log.Info().Msgf("%s: %.3fs", helper.GetMessage("total_inference_time", lang), total)
	log.Info().Msg(strings.Repeat("=", 50))
}
