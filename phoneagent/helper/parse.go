package helper

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	logs "github.com/sirupsen/logrus"
)

type Action map[string]any

type ActionResult struct {
	Success              bool
	ShouldFinish         bool
	Message              string
	RequiresConfirmation bool
}

// ParseFunctionCall converts OpenAI function call to Action format
func ParseFunctionCall(toolCall openai.ToolCall) (Action, error) {
	logs.Debugf("begin to parse function call: %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)

	action := Action{
		"_metadata": "do",
	}

	// Parse function arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, fmt.Errorf("failed to parse function arguments: %w", err)
	}

	// Map function name to action name
	actionName, err := mapFunctionToAction(toolCall.Function.Name)
	if err != nil {
		return nil, err
	}

	// Handle finish task specially
	if actionName == "finish" {
		action["_metadata"] = "finish"
		if msg, ok := args["message"].(string); ok {
			action["message"] = msg
		} else {
			action["message"] = "Task completed"
		}
		return action, nil
	}

	// Set action name
	action["action"] = actionName

	// Copy all arguments
	for k, v := range args {
		// Convert array coordinates from float64 to int
		if k == "element" || k == "start" || k == "end" {
			if arr, ok := v.([]interface{}); ok {
				intArr := make([]int, len(arr))
				for i, val := range arr {
					if fval, ok := val.(float64); ok {
						intArr[i] = int(fval)
					}
				}
				action[k] = intArr
				continue
			}
		}
		action[k] = v
	}

	return action, nil
}

// mapFunctionToAction maps function call names to internal action names
func mapFunctionToAction(funcName string) (string, error) {
	mapping := map[string]string{
		"tap":          "Tap",
		"type_text":    "Type",
		"swipe":        "Swipe",
		"long_press":   "Long Press",
		"double_tap":   "Double Tap",
		"launch_app":   "Launch",
		"press_back":   "Back",
		"press_home":   "Home",
		"wait":         "Wait",
		"take_over":    "Take_over",
		"interact":     "Interact",
		"record_note":  "Note",
		"call_api":     "Call_API",
		"finish_task":  "finish",
	}

	if actionName, ok := mapping[funcName]; ok {
		return actionName, nil
	}

	return "", fmt.Errorf("unknown function name: %s", funcName)
}
