package phoneagent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spance/autoglm-go/constants"
	"github.com/spance/autoglm-go/phoneagent/definitions"
	"github.com/spance/autoglm-go/phoneagent/helper"
	"github.com/spance/autoglm-go/phoneagent/llm"
	"github.com/spance/autoglm-go/utils"
)

type PhoneAgent struct {
	Device      Device
	ModelConfig *definitions.ModelConfig
	AgentConfig *definitions.AgentConfig
	State       []openai.ChatCompletionMessage
	StepCount   int
	ModelClient *llm.ModelClient
}

func NewPhoneAgent(device Device, modelConfig *definitions.ModelConfig, agentConfig *definitions.AgentConfig) *PhoneAgent {
	agentConfig.InitSystemPrompt()
	result := &PhoneAgent{
		ModelConfig: modelConfig,
		AgentConfig: agentConfig,
		State:       []openai.ChatCompletionMessage{},
		StepCount:   0,
		Device:      device,
		ModelClient: llm.NewModelClient(modelConfig),
	}
	return result
}

type StepResult struct {
	Success  bool
	Finished bool
	Action   map[string]interface{}
	Thinking string
	Message  string
}

func (r *PhoneAgent) Run(ctx context.Context, task string) (string, error) {
	result, err := r.ExecuteStep(ctx, task, true)
	if err != nil {
		log.Error().Int("step", r.StepCount).Err(err).Msg("Failed to execute step")
		return "", err
	}
	if result.Finished {
		return result.Message, nil
	}
	// Continue until finished or max steps reached
	for r.StepCount < r.AgentConfig.MaxSteps {
		result, err = r.ExecuteStep(ctx, "", false)
		if err != nil {
			log.Error().Int("step", r.StepCount).Err(err).Msg("Failed to execute step")
			return "", err
		}
		if result.Finished {
			return result.Message, nil
		}
	}
	return "Max steps reached", nil
}

func (r *PhoneAgent) Step(ctx context.Context, task string) (*StepResult, error) {
	isFirst := len(r.State) == 0
	if isFirst && len(task) == 0 {
		log.Error().Msg("task is required for the first step")
		return nil, fmt.Errorf("task is required for the first step")
	}
	return r.ExecuteStep(ctx, task, isFirst)
}

func (r *PhoneAgent) ExecuteStep(ctx context.Context, userPrompt string, isFirstStep bool) (*StepResult, error) {
	r.StepCount += 1

	device := r.Device
	screenshot, err := device.GetScreenshot(ctx, r.AgentConfig.DeviceID)
	if err != nil {
		log.Error().Int("step", r.StepCount).Err(err).Msg("Failed to get screenshot")
		return &StepResult{
			Success:  false,
			Finished: false,
			Message:  fmt.Sprintf("Failed to get screenshot: %v", err),
		}, err
	}

	currentApp, err := device.GetCurrentApp(ctx, r.AgentConfig.DeviceID)
	if err != nil {
		log.Warn().Int("step", r.StepCount).Err(err).Msg("Failed to get current app, continuing anyway")
		currentApp = "" // Use empty string as fallback
	}

	var textContent string
	if isFirstStep {
		// system prompt
		r.State = append(r.State,
			helper.CreateSystemMessage(r.AgentConfig.GetSystemPrompt()),
		)

		if len(currentApp) > 0 {
			screenInfo := helper.BuildScreenInfo(currentApp, screenshot)
			textContent = fmt.Sprintf("%s\n\n%s", userPrompt, screenInfo)
		} else {
			textContent = userPrompt
		}
	} else {
		var sb strings.Builder
		if len(userPrompt) > 0 {
			sb.WriteString(userPrompt)
			sb.WriteString("\n\n")
		}
		if len(currentApp) > 0 {
			screenInfo := helper.BuildScreenInfo(currentApp, screenshot)
			sb.WriteString("** Screen Info **\n\n")
			sb.WriteString(screenInfo)
		}
		textContent = sb.String()
	}

	// user prompt
	r.State = append(r.State,
		helper.CreateUserMessage(textContent, screenshot),
	)

	// print user message
	helper.PrintChatMessage(&r.State[len(r.State)-1], r.StepCount)

	response, err := r.ModelClient.Request(ctx, r.State)
	if err != nil {
		log.Error().Int("step", r.StepCount).Err(err).Msg("failed to get model response")
		return &StepResult{
			Success:  false,
			Finished: false,
			Message:  fmt.Sprintf("failed to get model response, err: %v", err),
		}, nil
	}

	log.Trace().Str("response", utils.JsonString(response)).Msg("💭 model response")

	// Parse action from function call
	var action helper.Action
	if len(response.ToolCalls) > 0 {
		action, err = helper.ParseFunctionCall(response.ToolCalls[0])
		if err != nil {
			log.Error().Int("step", r.StepCount).Err(err).Msg("failed to parse function call")
			return &StepResult{
				Success:  false,
				Finished: false,
				Message:  fmt.Sprintf("failed to parse function call, err: %v", err),
			}, nil
		}
	} else {
		// No tool call, might be a thinking step or error
		log.Warn().Int("step", r.StepCount).Msg("No tool call in response")
		return &StepResult{
			Success:  false,
			Finished: false,
			Message:  "Model did not return a tool call",
		}, nil
	}

	// Print action
	log.Debug().Int("step", r.StepCount).Str("action", response.Action).Str("details", utils.JsonString(action)).Msg("parsed action")

	// Remove image from context to save space
	helper.RemoveImagesFromMessage(&r.State[len(r.State)-1])

	// Add assistant message to state (including tool call)
	assistantMsg := openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		Content:   response.Thinking,
		ToolCalls: response.ToolCalls,
	}
	r.State = append(r.State, assistantMsg)

	// Execute action
	actionResult, err := r.ExecuteAction(ctx, action, screenshot.Width, screenshot.Height)
	if err != nil {
		log.Error().Int("step", r.StepCount).Err(err).Msg("failed to execute action")
		actionResult = helper.ActionResult{
			Success:      true,
			ShouldFinish: false,
			Message:      fmt.Sprintf("Action execution error: %v", err),
		}
	}

	// Add tool response message to state
	if len(response.ToolCalls) > 0 {
		toolMsg := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    actionResult.Message,
			ToolCallID: response.ToolCalls[0].ID,
		}
		r.State = append(r.State, toolMsg)
	}

	if actionResult.ShouldFinish {
		var displayMsg string
		if actionResult.Message != "" {
			displayMsg = actionResult.Message
		} else {
			displayMsg = helper.GetMessage("done", r.AgentConfig.Lang)
		}

		log.Debug().Int("step", r.StepCount).Msgf("✅ %s: %s", helper.GetMessage("task_completed", r.AgentConfig.Lang), displayMsg)
	}

	stepResult := &StepResult{
		Success:  actionResult.Success,
		Finished: actionResult.ShouldFinish,
		Action:   action,
		Thinking: response.Thinking,
	}
	if len(actionResult.Message) > 0 {
		stepResult.Message = actionResult.Message
	} else {
		stepResult.Message = utils.AnyToString(action["message"])
	}

	return stepResult, nil
}

func (r *PhoneAgent) ExecuteAction(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	actionType := utils.AnyToString(action["_metadata"])

	if actionType == "finish" {
		return helper.ActionResult{
			Success:      true,
			ShouldFinish: true,
			Message:      utils.AnyToString(action["message"]),
		}, nil
	}
	if actionType != "do" {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      fmt.Sprintf("Unknown action type: %s", actionType),
		}, nil
	}

	actionName := utils.AnyToString(action["action"])
	switch actionName {
	case "Launch":
		return r.handleLaunch(ctx, action, screenWidth, screenHeight)
	case "Tap":
		return r.handleTap(ctx, action, screenWidth, screenHeight)
	case "Type":
		return r.handleType(ctx, action, screenWidth, screenHeight)
	case "Type_Name":
		return r.handleType(ctx, action, screenWidth, screenHeight)
	case "Swipe":
		return r.handleSwipe(ctx, action, screenWidth, screenHeight)
	case "Back":
		return r.handleBack(ctx, action, screenWidth, screenHeight)
	case "Home":
		return r.handleHome(ctx, action, screenWidth, screenHeight)
	case "Double Tap":
		return r.handleDoubleTap(ctx, action, screenWidth, screenHeight)
	case "Long Press":
		return r.handleLongPress(ctx, action, screenWidth, screenHeight)
	case "Wait":
		return r.handleWait(ctx, action, screenWidth, screenHeight)
	case "Take_over":
		return r.handleTakeover(ctx, action, screenWidth, screenHeight)
	case "Note":
		return r.handleNote(ctx, action, screenWidth, screenHeight)
	case "Call_API":
		return r.handleCallAPI(ctx, action, screenWidth, screenHeight)
	case "Interact":
		return r.handleInteract(ctx, action, screenWidth, screenHeight)
	default:
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      fmt.Sprintf("Unknown action name: %s", actionName),
		}, nil
	}
}

func (r *PhoneAgent) handleLaunch(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	appName := utils.AnyToString(action["app"])
	if len(appName) == 0 {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      "No app name specified",
		}, nil
	}

	packageName, ok := constants.GetPackageByAlias(appName)
	if !ok || len(packageName) == 0 {
		packageName = appName // assume it's a package name
	}

	_, err := r.Device.LaunchApp(ctx, packageName, r.AgentConfig.DeviceID)
	if err != nil {
		log.Error().Int("step", r.StepCount).Err(err).Msg("failed to launch app")
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      fmt.Sprintf("failed to launch app, err: %v", err),
		}, nil
	}

	return helper.ActionResult{
		Success:      true,
		ShouldFinish: false,
	}, nil
}

func (r *PhoneAgent) convertRelativeToAbsolute(element []int, screenWidth, screenHeight int) (int, int) {
	x := int(float64(element[0]) / float64(1000) * float64(screenWidth))
	y := int(float64(element[1]) / float64(1000) * float64(screenHeight))
	return x, y
}

func (r *PhoneAgent) DefaultConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Sensitive operation: %s\nConfirm? (Y/N): ", message)

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	response = strings.ToUpper(response)

	return response == "Y"
}

func (r *PhoneAgent) DefaultTakeover(message string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\nPress Enter after completing manual operation...", message)
	_, _ = reader.ReadString('\n')
}

func (r *PhoneAgent) handleTap(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	element := utils.AnyToIntSlice(action["element"])
	if len(element) != 2 {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      "Invalid element coordinates",
		}, nil
	}

	x, y := r.convertRelativeToAbsolute(element, screenWidth, screenHeight)
	if msg, ok := action["message"]; ok {
		if !r.DefaultConfirmation(utils.AnyToString(msg)) {
			return helper.ActionResult{
				Success:      false,
				ShouldFinish: true,
				Message:      "User cancelled sensitive operation",
			}, nil
		}
	}
	_ = r.Device.Tap(ctx, x, y, r.AgentConfig.DeviceID)

	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) Reset(ctx context.Context) {
	r.State = []openai.ChatCompletionMessage{}
	r.StepCount = 0
}

func (r *PhoneAgent) handleType(ctx context.Context, action helper.Action, width int, height int) (helper.ActionResult, error) {
	text := utils.AnyToString(action["text"])
	device := r.Device
	deviceID := r.AgentConfig.DeviceID

	// Switch to ADB keyboard
	originalIME, _ := device.DetectAndSetADBKeyboard(ctx, deviceID)
	time.Sleep(time.Second * 1)

	// Clear existing text and type new text
	_ = device.ClearText(ctx, deviceID)
	time.Sleep(time.Second * 1)

	// Handle multiline text by splitting on newlines
	_ = device.TypeText(ctx, text, deviceID)
	time.Sleep(time.Second * 1)

	// Restore original keyboard
	_ = device.RestoreKeyboard(ctx, originalIME, deviceID)
	time.Sleep(time.Second * 1)

	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleSwipe(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	start := utils.AnyToIntSlice(action["start"])
	end := utils.AnyToIntSlice(action["end"])
	if len(start) != 2 || len(end) != 2 {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: false,
			Message:      "Invalid swipe coordinates",
		}, nil
	}
	startX, startY := r.convertRelativeToAbsolute(start, screenWidth, screenHeight)
	endX, endY := r.convertRelativeToAbsolute(end, screenWidth, screenHeight)
	_ = r.Device.Swipe(ctx, startX, startY, endX, endY, r.AgentConfig.DeviceID)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleBack(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	_ = r.Device.Back(ctx, r.AgentConfig.DeviceID)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleHome(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	_ = r.Device.Home(ctx, r.AgentConfig.DeviceID)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleDoubleTap(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	element := utils.AnyToIntSlice(action["element"])
	if len(element) != 2 {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: true,
			Message:      "Invalid element coordinates",
		}, nil
	}
	x, y := r.convertRelativeToAbsolute(element, screenWidth, screenHeight)
	_ = r.Device.DoubleTap(ctx, x, y, r.AgentConfig.DeviceID)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleLongPress(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	element := utils.AnyToIntSlice(action["element"])
	if len(element) != 2 {
		return helper.ActionResult{
			Success:      false,
			ShouldFinish: true,
			Message:      "Invalid element coordinates",
		}, nil
	}
	x, y := r.convertRelativeToAbsolute(element, screenWidth, screenHeight)
	_ = r.Device.LongPress(ctx, x, y, r.AgentConfig.DeviceID)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleWait(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	durationStr := utils.AnyToString(action["duration"])
	duration, err := strconv.ParseFloat(strings.ReplaceAll(durationStr, "seconds", ""), 64)
	if err != nil {
		log.Warn().Int("step", r.StepCount).Err(err).Msg("failed to parse duration, using default 1.0s")
		duration = 1.0
	}
	time.Sleep(time.Duration(duration) * time.Second)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleTakeover(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	message := utils.AnyToString(action["message"])
	if message == "" {
		message = "User intervention required"
	}
	r.DefaultTakeover(message)
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleNote(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	// This action is typically used for recording page content
	// Implementation depends on specific requirements
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleCallAPI(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	// This action is typically used for content summarization
	// Implementation depends on specific requirements
	return helper.ActionResult{Success: true, ShouldFinish: false}, nil
}

func (r *PhoneAgent) handleInteract(ctx context.Context, action helper.Action, screenWidth, screenHeight int) (helper.ActionResult, error) {
	// This action signals that user input is needed
	return helper.ActionResult{Success: true, ShouldFinish: false, Message: "User interaction required"}, nil
}
