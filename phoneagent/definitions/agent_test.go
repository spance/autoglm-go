package definitions

import (
	"strings"
	"testing"
)

func TestGetSystemPrompt(t *testing.T) {
	// Test case 1: Lang = "en"
	config1 := &AgentConfig{
		Lang: "en",
	}
	config1.InitSystemPrompt()
	result1 := config1.GetSystemPrompt()
	if !strings.Contains(result1, "The current date:") {
		t.Errorf("Expected English prompt, got: %s", result1)
	}
	if !strings.Contains(result1, "Android operation agent") {
		t.Errorf("Expected English content in prompt")
	}
	t.Logf("=== Lang = 'en' ===\n%s\n", result1)

	// Test case 2: Lang = "cn"
	config2 := &AgentConfig{
		Lang: "cn",
	}
	config2.InitSystemPrompt()
	result2 := config2.GetSystemPrompt()
	if !strings.Contains(result2, "今天的日期是:") {
		t.Errorf("Expected Chinese prompt, got: %s", result2)
	}
	if !strings.Contains(result2, "智能手机操作助手") {
		t.Errorf("Expected Chinese content in prompt")
	}
	t.Logf("=== Lang = 'cn' ===\n%s\n", result2)

	// Test case 3: Lang = other (should default to Chinese)
	config3 := &AgentConfig{
		Lang: "other",
	}
	config3.InitSystemPrompt()
	result3 := config3.GetSystemPrompt()
	if !strings.Contains(result3, "今天的日期是:") {
		t.Errorf("Expected Chinese prompt for unknown language, got: %s", result3)
	}
	t.Logf("=== Lang = 'other' ===\n%s\n", result3)
}
