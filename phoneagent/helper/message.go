package helper

import (
	"bytes"
	"encoding/base64"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spance/autoglm-go/constants"
	"github.com/spance/autoglm-go/phoneagent/definitions"
	"github.com/spance/autoglm-go/utils"
)

func CreateSystemMessage(content string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: content,
	}
}

func CreateAssistantMessage(content string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	}
}

func CreateUserMessage(text string, screenshot *definitions.Screenshot) openai.ChatCompletionMessage {
	msg := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{
			{
				Type: openai.ChatMessagePartTypeText,
				Text: text,
			},
		},
	}

	var imageURL bytes.Buffer
	imageURL.WriteString("data:image/png;base64,")

	// 优先使用二进制数据，其次使用 Base64Data
	if len(screenshot.BinaryData) > 0 {
		encoder := base64.NewEncoder(base64.StdEncoding, &imageURL)
		encoder.Write(screenshot.BinaryData)
		encoder.Close()
	} else if screenshot.Base64Data != "" {
		// 使用 Base64 数据
		imageURL.WriteString(screenshot.Base64Data)
	}

	if imageURL.Len() > 0 {
		msg.MultiContent = append(msg.MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: imageURL.String(),
			},
		})
	}

	return msg
}

func PrintChatMessage(msg *openai.ChatCompletionMessage, stepCount int) {
	// user 只打印 text
	if msg.Role == openai.ChatMessageRoleUser {
		for _, part := range msg.MultiContent {
			if part.Type == openai.ChatMessagePartTypeText {
				log.Debug().Int("step", stepCount).Str("Text", part.Text).Msg("👤 user message")
			}
		}
	}
	// assistant 打印 content
	if msg.Role == openai.ChatMessageRoleAssistant {
		log.Debug().Int("step", stepCount).Str("Content", msg.Content).Msg("🌐 assistant message")
	}
	// 不打印 system prompt
	if msg.Role == openai.ChatMessageRoleSystem {
		return
	}
}

func BuildScreenInfo(currentApp string) string {
	appName, _ := constants.GetAliasByPackage(currentApp)
	info := map[string]any{
		"current_app":      currentApp,
		"current_app_name": appName,
	}
	return utils.JsonString(info)
}

func GetMessage(key string, lang string) string {
	if lang == "en" {
		return constants.MESSAGES_EN_MAP[key]
	}
	return constants.MESSAGES_ZH_MAP[key]
}

func RemoveImagesFromMessage(message *openai.ChatCompletionMessage) {
	if message == nil || message.MultiContent == nil {
		return
	}
	var multiContent []openai.ChatMessagePart
	for _, part := range message.MultiContent {
		// Remove image URLs, keep only text parts
		if part.Type != openai.ChatMessagePartTypeImageURL {
			multiContent = append(multiContent, part)
		}
	}
	message.MultiContent = multiContent
}
