package llm

import (
	"autoglm-go/phoneagent/definitions"
	"autoglm-go/phoneagent/helper"
	"autoglm-go/utils"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"strings"
	"time"
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

		rawContent         strings.Builder
		buffer             strings.Builder
		inActionPhase      bool
		firstTokenReceived bool
	)

	req := openai.ChatCompletionRequest{
		Model:               c.config.ModelName,
		Messages:            messages,
		MaxCompletionTokens: c.config.MaxTokens,
		Temperature:         c.config.Temperature,
		TopP:                c.config.TopP,
		FrequencyPenalty:    c.config.FrequencyPenalty,
		Stream:              true,
	}

	for _, data := range req.Messages[len(req.Messages)-1].MultiContent {
		if data.Text != "" {
			fmt.Printf("data: %s\n", utils.JsonString(data))
		}
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("CreateChatCompletionStream error: %v\n", err)
		return nil, err
	}
	defer stream.Close()

	actionMarkers := []string{"finish(message=", "do(action="}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Stream error: %v\n", err)
			return nil, err
		}

		if len(resp.Choices) == 0 {
			continue
		}

		delta := resp.Choices[0].Delta.Content
		if delta == "" {
			continue
		}

		rawContent.WriteString(delta)

		// time to first token
		if !firstTokenReceived {
			t := time.Since(startTime).Seconds()
			timeToFirstToken = &t
			firstTokenReceived = true
		}

		if inActionPhase {
			continue
		}

		buffer.WriteString(delta)
		bufStr := buffer.String()

		markerFound := false
		for _, marker := range actionMarkers {
			if strings.Contains(bufStr, marker) {
				thinkingPart := strings.SplitN(bufStr, marker, 2)[0]
				fmt.Print(thinkingPart)
				fmt.Println()

				inActionPhase = true
				markerFound = true

				if timeToThinkingEnd == nil {
					t := time.Since(startTime).Seconds()
					timeToThinkingEnd = &t
				}
				break
			}
		}

		if markerFound {
			continue
		}

		// Check if buffer ends with a prefix of any marker
		// If so, don't print yet (wait for more content)
		isPotentialMarker := false
		for _, marker := range actionMarkers {
			for i := 1; i < len(marker); i++ {
				if strings.HasSuffix(bufStr, marker[:i]) {
					isPotentialMarker = true
					break
				}
			}
			if isPotentialMarker {
				break
			}
		}

		if !isPotentialMarker {
			// Safe to print the buffer
			fmt.Print(bufStr)
			buffer.Reset()
		}
	}

	totalTime := time.Since(startTime).Seconds()

	thinking, action := parseResponse(rawContent.String())

	printMetrics(
		c.config.Lang,
		timeToFirstToken,
		timeToThinkingEnd,
		totalTime,
	)

	return &ModelResponse{
		Thinking:          thinking,
		Action:            action,
		RawContent:        rawContent.String(),
		TimeToFirstToken:  timeToFirstToken,
		TimeToThinkingEnd: timeToThinkingEnd,
		TotalTime:         totalTime,
	}, nil
}
func parseResponse(content string) (string, string) {
	/*
	   Parse the model response into thinking and action parts.

	   Parsing rules:
	   1. If content contains 'finish(message=', everything before is thinking,
	      everything from 'finish(message=' onwards is action.
	   2. If rule 1 doesn't apply but content contains 'do(action=',
	      everything before is thinking, everything from 'do(action=' onwards is action.
	   3. Fallback: If content contains '<answer>', use legacy parsing with XML tags.
	   4. Otherwise, return empty thinking and full content as action.

	   Args:
	       content: Raw response content.

	   Returns:
	       Tuple of (thinking, action).
	*/

	// Rule 1: Check for finish(message=
	if strings.Contains(content, "finish(message=") {
		parts := strings.SplitN(content, "finish(message=", 2)
		return strings.TrimSpace(parts[0]), "finish(message=" + parts[1]
	}

	// Rule 2: Check for do(action=
	if strings.Contains(content, "do(action=") {
		parts := strings.SplitN(content, "do(action=", 2)
		return strings.TrimSpace(parts[0]), "do(action=" + parts[1]
	}

	// Rule 3: Fallback to legacy XML tag parsing
	if strings.Contains(content, "<answer>") {
		parts := strings.SplitN(content, "<answer>", 2)
		thinking := strings.TrimSpace(
			strings.ReplaceAll(
				strings.ReplaceAll(parts[0], "<think>", ""),
				"</think>", "",
			),
		)
		action := strings.TrimSpace(
			strings.ReplaceAll(parts[1], "</answer>", ""),
		)
		return thinking, action
	}

	// Rule 4: No markers found, return content as action
	return "", content
}

func printMetrics(lang string, firstToken *float64, thinkingEnd *float64, total float64) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("⏱️  " + helper.GetMessage("performance_metrics", lang))
	fmt.Println(strings.Repeat("-", 50))

	if firstToken != nil {
		fmt.Printf("%s: %.3fs\n", helper.GetMessage("time_to_first_token", lang), *firstToken)
	}
	if thinkingEnd != nil {
		fmt.Printf("%s: %.3fs\n", helper.GetMessage("time_to_thinking_end", lang), *thinkingEnd)
	}
	fmt.Printf("%s: %.3fs\n", helper.GetMessage("total_inference_time", lang), total)
	fmt.Println(strings.Repeat("=", 50))
}
