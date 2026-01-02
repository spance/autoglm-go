package helper

import (
	"autoglm-go/constants"
	"autoglm-go/utils"
	"fmt"
	"github.com/sashabaranov/go-openai"
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

func CreateUserMessage(text string, imageBase64 *string) openai.ChatCompletionMessage {
	msg := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{
			{
				Type: openai.ChatMessagePartTypeText,
				Text: text,
			},
		},
	}
	// 如果有图片，加入 MultiContent
	if imageBase64 != nil {
		msg.MultiContent = append(msg.MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: fmt.Sprintf("data:image/png;base64,%s", *imageBase64),
			},
		})
	}
	return msg
}

func BuildScreenInfo(currentApp string) string {
	info := map[string]interface{}{
		"current_app": currentApp,
	}
	return utils.JsonString(info)
}

func GetMessage(key string, lang string) string {
	if lang == "en" {
		return constants.MESSAGES_EN_MAP[key]
	}
	return constants.MESSAGES_ZH_MAP[key]
}

func RemoveImagesFromMessage(message openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	var multiContent []openai.ChatMessagePart
	if message.MultiContent != nil {
		for _, part := range message.MultiContent {
			if part.Type == openai.ChatMessagePartTypeText {
				multiContent = append(multiContent, part)
			}
		}
		message.MultiContent = multiContent
	}
	return message
}
