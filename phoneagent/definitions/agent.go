package definitions

import (
	"fmt"
	"os"
	"time"

	"github.com/spance/autoglm-go/constants"
	"github.com/valyala/fasttemplate"
)

// AgentConfig 代理配置
type AgentConfig struct {
	MaxSteps       int                    // 最大执行步数
	DeviceID       string                 // 设备 ID
	Lang           string                 // 语言设置: "en" 或 "cn"
	WdaUrl         string                 // WebDriverAgent URL (仅 iOS)
	PromptPath     string                 // 自定义系统提示文件路径（可选）
	promptTemplate *fasttemplate.Template // 缓存的提示模板
}

var (
	// weekdayNamesCN 中文星期名称，索引对应 time.Weekday (0=Sunday, 1=Monday, ...)
	weekdayNamesCN = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
)

// InitSystemPrompt 初始化系统提示模板
// 优先从自定义文件读取，否则使用内置模板
// 应在创建 AgentConfig 后立即调用
func (c *AgentConfig) InitSystemPrompt() error {
	var templateContent string

	// 优先级1: 从自定义文件加载
	if c.PromptPath != "" {
		if content, err := os.ReadFile(c.PromptPath); err == nil {
			templateContent = string(content)
		}
	}

	// 优先级2: 使用内置默认模板
	if templateContent == "" {
		if c.Lang == "en" {
			templateContent = constants.SystemPrompt_EN
		} else {
			templateContent = constants.SystemPrompt_ZH
		}
	}

	// 编译模板（使用 {{ }} 作为变量占位符）
	c.promptTemplate = fasttemplate.New(templateContent, "{{", "}}")
	return nil
}

// GetSystemPrompt 获取系统提示（含当前日期时间）
// 每次调用都会动态生成最新的日期时间信息
func (c *AgentConfig) GetSystemPrompt() string {
	if c.promptTemplate == nil {
		return ""
	}

	now := time.Now()
	dateTimeStr := c.formatDateTime(now)

	// 渲染模板并返回
	return c.promptTemplate.ExecuteString(map[string]any{
		"datetime": dateTimeStr,
	})
}

// formatDateTime 格式化日期时间为中英文格式
func (c *AgentConfig) formatDateTime(t time.Time) string {
	if c.Lang == "en" {
		// 英文格式: February 01, 2026, 15:04:05
		return t.Format("January 02, 2006, 15:04:05")
	}

	// 中文格式: 2026年2月1日 星期日 15:04:05
	weekday := weekdayNamesCN[t.Weekday()]
	return fmt.Sprintf("%d年%d月%d日 %s %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(), weekday,
		t.Hour(), t.Minute(), t.Second())
}
