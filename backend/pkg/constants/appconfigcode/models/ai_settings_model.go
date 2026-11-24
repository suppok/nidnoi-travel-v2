package models

import (
	"backend/pkg/constants/aimodel"
	"backend/pkg/core/validation"
)

// AppConfigAISettings is the value of AI_SETTINGS, edited on the backoffice page "ตั้งค่า AI".
// The Anthropic API key is stored only encrypted (AES-256-GCM with AI_SETTINGS_ENCRYPTION_KEY, the code
// as additional data) and never returned to a client; apiKeyLast4 is what the page shows.
type AppConfigAISettings struct {
	Enabled          bool   `json:"enabled"`
	Model            string `json:"model"`
	MaxPdfSizeMb     int    `json:"maxPdfSizeMb"`
	APIKeyCiphertext string `json:"apiKeyCiphertext"` // "" = no key
	APIKeyLast4      string `json:"apiKeyLast4"`
}

// MaxPdfSizeMbLimit is the highest value the setting accepts; config ai.maxPdfMB may cap it lower.
const MaxPdfSizeMbLimit = 90

func (m AppConfigAISettings) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Model, validation.Required, validation.In(aimodel.IDs()...)),
		validation.Field(&m.MaxPdfSizeMb, validation.Required, validation.Min(1), validation.Max(MaxPdfSizeMbLimit)),
		validation.Field(&m.APIKeyCiphertext, validation.RuneLength(0, 2000)),
		validation.Field(&m.APIKeyLast4, validation.RuneLength(0, 4)),
	)
}
