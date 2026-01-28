package definitions

import "github.com/sashabaranov/go-openai"

// Parameter definition helpers
type ParamProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Items       *ParamItems `json:"items,omitempty"`
	MinItems    *int        `json:"minItems,omitempty"`
	MaxItems    *int        `json:"maxItems,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

type ParamItems struct {
	Type string `json:"type"`
}

type FunctionParams struct {
	Type       string                   `json:"type"`
	Properties map[string]ParamProperty `json:"properties"`
	Required   []string                 `json:"required,omitempty"`
}

// Coordinate array parameter (reusable)
func coordinateParam(description string) ParamProperty {
	minItems := 2
	maxItems := 2
	return ParamProperty{
		Type:        "array",
		Description: description,
		Items:       &ParamItems{Type: "integer"},
		MinItems:    &minItems,
		MaxItems:    &maxItems,
	}
}

// String parameter (reusable)
func stringParam(description string, required bool) ParamProperty {
	return ParamProperty{
		Type:        "string",
		Description: description,
	}
}

// Number parameter (reusable)
func numberParam(description string, defaultValue float64) ParamProperty {
	return ParamProperty{
		Type:        "number",
		Description: description,
		Default:     defaultValue,
	}
}

// GetPhoneAgentTools returns all available function tools for the phone agent
func GetPhoneAgentTools() []openai.Tool {
	return []openai.Tool{
		createTapTool(),
		createTypeTextTool(),
		createSwipeTool(),
		createLongPressTool(),
		createDoubleTapTool(),
		createLaunchAppTool(),
		createPressBackTool(),
		createPressHomeTool(),
		createWaitTool(),
		createTakeOverTool(),
		createInteractTool(),
		createRecordNoteTool(),
		createCallAPITool(),
		createFinishTaskTool(),
	}
}

func createTapTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "tap",
			Description: "Perform a tap action on a specified screen area. Use this to click buttons, select items, open apps from home screen, or interact with any clickable UI elements.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"element": coordinateParam("Coordinates of the tap point [x, y]"),
					"message": stringParam("Optional message for sensitive operations (payments, privacy, etc.)", false),
				},
				Required: []string{"element"},
			},
		},
	}
}

func createTypeTextTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "type_text",
			Description: "Enter text into the currently focused input field. Make sure the input field is focused (tap it first) before using this action.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"text": stringParam("The text to input", true),
				},
				Required: []string{"text"},
			},
		},
	}
}

func createSwipeTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "swipe",
			Description: "Perform a swipe gesture by dragging from start coordinates to end coordinates. Use for scrolling content, navigating between screens, pulling down notifications, or gesture-based interactions.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"start": coordinateParam("Start coordinates [x1, y1]"),
					"end":   coordinateParam("End coordinates [x2, y2]"),
				},
				Required: []string{"start", "end"},
			},
		},
	}
}

func createLongPressTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "long_press",
			Description: "Perform a long press action on a specific screen point. Use for triggering context menus, selecting text, or activating long-press interactions.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"element": coordinateParam("Coordinates of the long press point [x, y]"),
				},
				Required: []string{"element"},
			},
		},
	}
}

func createDoubleTapTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "double_tap",
			Description: "Quickly tap twice on a specific screen point. Use for activating double-tap interactions like zoom, text selection, or opening items.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"element": coordinateParam("Coordinates of the double tap point [x, y]"),
				},
				Required: []string{"element"},
			},
		},
	}
}

func createLaunchAppTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "launch_app",
			Description: "Launch a target app. This is faster than navigating through the home screen. After this action completes, you will automatically receive a screenshot of the result.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"app": stringParam("Name of the app to launch", true),
				},
				Required: []string{"app"},
			},
		},
	}
}

func createPressBackTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "press_back",
			Description: "Navigate back to the previous screen or close the current dialog. Equivalent to pressing Android's back button. Use to return from deeper screens, close popups, or exit the current context.",
			Parameters: FunctionParams{
				Type:       "object",
				Properties: map[string]ParamProperty{},
			},
		},
	}
}

func createPressHomeTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "press_home",
			Description: "Return to the system home screen. Equivalent to pressing Android's home button. Use to exit the current app and return to launcher, or start a new task from a known state.",
			Parameters: FunctionParams{
				Type:       "object",
				Properties: map[string]ParamProperty{},
			},
		},
	}
}

func createWaitTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "wait",
			Description: "Wait for page to load. Use when content is loading or animations are in progress.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"duration": numberParam("Duration in seconds to wait", 1.0),
				},
			},
		},
	}
}

func createTakeOverTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "take_over",
			Description: "Request user assistance for login, verification, or other manual operations that the agent cannot complete automatically.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"message": stringParam("Message explaining what user needs to do", true),
				},
				Required: []string{"message"},
			},
		},
	}
}

func createInteractTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "interact",
			Description: "Request user interaction when there are multiple options that meet the criteria and user needs to choose.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"message": stringParam("Description of the interaction needed", false),
				},
			},
		},
	}
}

func createRecordNoteTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "record_note",
			Description: "Record current page content for later summarization.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"message": stringParam("Content to record", false),
				},
			},
		},
	}
}

func createCallAPITool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "call_api",
			Description: "Summarize or comment on current page or recorded content.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"instruction": stringParam("Instruction for what to summarize or analyze", true),
				},
				Required: []string{"instruction"},
			},
		},
	}
}

func createFinishTaskTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "finish_task",
			Description: "Complete the task successfully. Use this when the task has been accomplished accurately and completely.",
			Parameters: FunctionParams{
				Type: "object",
				Properties: map[string]ParamProperty{
					"message": stringParam("Completion message explaining what was accomplished", true),
				},
				Required: []string{"message"},
			},
		},
	}
}
