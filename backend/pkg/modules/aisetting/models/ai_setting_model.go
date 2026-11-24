package models

import (
	"github.com/guregu/null/v5"

	"backend/pkg/constants/aimodel"
	coremodels "backend/pkg/core/api/models"
)

// --- get ---

type AISettingGetRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// AISettingResponse never carries the API key itself: only whether one is stored and its last 4 characters.
type AISettingResponse struct {
	Enabled      bool   `json:"enabled"`
	Model        string `json:"model"`
	MaxPdfSizeMb int    `json:"maxPdfSizeMb"`
	HasAPIKey    bool   `json:"hasApiKey"`
	APIKeyLast4  string `json:"apiKeyLast4"` // "" when no key
	// KeyUnreadable: a key is stored but can't be decrypted (AI_SETTINGS_ENCRYPTION_KEY changed); enter it again.
	KeyUnreadable bool `json:"keyUnreadable"`
	// EncryptionConfigured is false while AI_SETTINGS_ENCRYPTION_KEY is not set on the server (AI stays off).
	EncryptionConfigured bool `json:"encryptionConfigured"`
	// Ready: enabled, a usable key and encryption configured — the AI prefill is offered to staff.
	Ready             bool            `json:"ready"`
	MaxPdfSizeMbLimit int             `json:"maxPdfSizeMbLimit"`               // highest maxPdfSizeMb the server accepts
	Models            []aimodel.Model `json:"models"`                          // selectable models, default first
	UpdateDate        null.Time       `json:"updateDate" swaggertype:"string"` // null until an admin saves
	UpdateBy          string          `json:"updateBy"`                        // name; "" for none
}

// --- update ---

// AISettingUpdateRequest replaces the settings. apiKey omitted, null or "" keeps the stored key;
// clearApiKey removes it (and AI can't stay enabled without one).
type AISettingUpdateRequest struct {
	Enabled      bool        `json:"enabled"`
	Model        string      `json:"model"`
	MaxPdfSizeMb int         `json:"maxPdfSizeMb"`
	APIKey       null.String `json:"apiKey" swaggertype:"string"` // a new Anthropic API key (sk-ant-…)
	ClearAPIKey  bool        `json:"clearApiKey"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// --- test ---

// AISettingTestRequest checks a key against Anthropic without saving it. Empty apiKey = the stored
// key; empty model = the stored model.
type AISettingTestRequest struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AISettingTestResponse struct {
	OK    bool   `json:"ok"`
	Model string `json:"model"` // the model that was checked
}
