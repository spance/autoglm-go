# AutoGLM-Go

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A production-ready Go library for AI-driven Android device automation. Programmatically execute complex user interactions on Android devices through natural language instructions via LLM function calling.

This project is forked from [ZoroSpace/autoglm-go](https://github.com/ZoroSpace/autoglm-go), which originated from [zai-org/Open-AutoGLM](https://github.com/zai-org/Open-AutoGLM).

## Overview

AutoGLM-Go provides a complete abstraction layer for autonomous Android device control using vision-language models. The library handles:

- Screenshot capture and vision processing
- AI-powered action planning from natural language
- Atomic device operations (tap, swipe, type, etc.)
- Stateful conversation management for multi-step task execution
- Structured logging with step tracking
- Both local ADB and remote device connections

## Installation

```bash
go get github.com/spance/autoglm-go
```

## Core Architecture

### Interfaces

```go
type Device interface {
    DeviceOperator  // Screenshot, touch, input operations
    DeviceManager   // Connection management, device enumeration
}

type DeviceOperator interface {
    GetScreenshot(ctx context.Context, deviceID string) (*definitions.Screenshot, error)
    GetCurrentApp(ctx context.Context, deviceID string) (string, error)
    Tap(ctx context.Context, x, y int, deviceID string) error
    Swipe(ctx context.Context, startX, startY, endX, endY int, deviceID string) error
    TypeText(ctx context.Context, text, deviceID string) error
    LaunchApp(ctx context.Context, appName, deviceID string) (bool, error)
    // ... additional operations
}
```

### Agent Loop

```go
agent := phoneagent.NewPhoneAgent(device, modelConfig, agentConfig)
result, err := agent.Run(ctx, "search for iPhone 15 and add to cart")
```

The agent implements an iterative cycle:
1. Capture device screenshot
2. Send screenshot + conversation history to LLM
3. Parse LLM function calls into structured actions
4. Execute actions via device interface
5. Repeat until task completion or max steps reached

## Usage

### Basic Integration

```go
import "github.com/spance/autoglm-go/phoneagent"

// Create device instance
device := &android.ADBDevice{}

// Configure LLM
modelConfig := &definitions.ModelConfig{
    BaseURL: "https://api.openai.com/v1",
    Model:   "gpt-4-vision",
    APIKey:  os.Getenv("OPENAI_API_KEY"),
}

// Configure agent behavior
agentConfig := &definitions.AgentConfig{
    DeviceID:  "emulator-5554",
    MaxSteps:  100,
    Lang:      "en",
}

// Create and run agent
agent := phoneagent.NewPhoneAgent(device, modelConfig, agentConfig)
result, err := agent.Run(ctx, "your task description")
```

### Device Management

```go
// Connect to remote device via TCP/IP
_, err := device.Connect(ctx, "192.168.1.100:5555")

// List connected devices
devices, err := device.ListDevices(ctx)

// Get device info
info, err := device.GetDeviceInfo(ctx, "device-id")
```

### Step-by-Step Execution

For fine-grained control:

```go
// Single execution step
result, err := agent.Step(ctx, "initial task prompt")

// Inspect result
if result.Finished {
    fmt.Println("Task completed:", result.Message)
}

// Continue with follow-up steps
result, err := agent.Step(ctx, "")
```

## Configuration

### Model Configuration

```go
type ModelConfig struct {
    BaseURL string  // LLM API endpoint
    Model   string  // Model identifier
    APIKey  string  // Authentication token
}
```

Supports OpenAI-compatible APIs. Tested with:
- OpenAI GPT-4V
- Claude Opus (via OpenAI-compatible proxy)
- Custom LLM servers with compatible API

### Agent Configuration

```go
type AgentConfig struct {
    DeviceID  string
    MaxSteps  int    // Max iterations per task
    Lang      string // "en" or "cn" for system prompts
}
```

## Coordinate System

All coordinates use normalized 0-999 range regardless of actual screen resolution. The library automatically converts to absolute device pixels:

```
(0, 0) ----------- (999, 0)
  |                   |
  |   normalized      |
  |   0-999 range     |
  |                   |
(0, 999) -------- (999, 999)
```

This abstraction is transparent to users - provide coordinates in 0-999 range, the library handles conversion.

## Structured Logging

Logging uses [zerolog](https://github.com/rs/zerolog) for structured output with step tracking:

```
log.Debug().Int("step", 1).Msgf("💭 thinking")
log.Debug().Int("step", 1).Msgf("🎯 parsed action: Tap")
log.Error().Int("step", 2).Err(err).Msg("failed to execute action")
```

## Extending

### Custom Device Implementation

Implement the `Device` interface to support additional platforms:

```go
type CustomDevice struct {
    // your implementation
}

func (d *CustomDevice) GetScreenshot(ctx context.Context, deviceID string) (*definitions.Screenshot, error) {
    // implementation
}

// ... implement remaining interface methods
```

### Custom LLM Models

The library uses OpenAI-compatible APIs. Any model exposing that interface is supported:

- Modify `modelConfig.BaseURL` to point to your LLM endpoint
- Ensure function calling is supported by the model
- Update system prompts in `constants/prompt.go` if needed

## Example: App Automation

See [examples/](./examples) for reference implementations including Android ADB device control.

## Limitations

- Android device support via ADB only
- Requires target device to have USB debugging enabled
- LLM must support vision input and function calling
- Performance depends on LLM response latency and device screenshot speed

## Performance Considerations

- Screenshot capture: ~100-300ms per device
- LLM inference: Model-dependent (typically 1-5s for vision models)
- Action execution: ~50-200ms per operation
- Overall task time: linear in number of steps required

## License

Apache License 2.0

See [LICENSE](LICENSE) for details.

## Acknowledgments

- [zai-org/Open-AutoGLM](https://github.com/zai-org/Open-AutoGLM) - Original project
- [ZoroSpace/autoglm-go](https://github.com/ZoroSpace/autoglm-go) - Previous Go implementation
