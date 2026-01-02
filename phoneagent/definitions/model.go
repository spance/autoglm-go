package definitions

type ModelConfig struct {
	BaseURL   string
	ModelName string
	APIKey    string
	Lang      string

	MaxTokens        int
	Temperature      float32
	TopP             float32
	FrequencyPenalty float32
}
