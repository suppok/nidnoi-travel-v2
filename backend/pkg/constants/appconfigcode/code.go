// Package appconfigcode registers every app config code with its value type.
// Read a value with appconfigload.GetAppConfig(ctx, repo, appconfigcode.X).
package appconfigcode

import (
	"backend/pkg/constants/appconfigcode/core"
	"backend/pkg/constants/appconfigcode/models"
)

// AppConfig is a registered code.
type AppConfig struct {
	AppConfig core.AppConfig
	// Required codes must have an active, valid row; POST /app-configs/verify reports them otherwise.
	Required bool
	// Features lists the features that need this config (informational).
	Features []string
}

var appConfigMap = map[string]*AppConfig{}

// GetAppConfigMap returns every registered code, keyed by code.
func GetAppConfigMap() map[string]*AppConfig { return appConfigMap }

// NewAppConfig declares a code. Declare codes as package-level vars below so the map lists them.
func NewAppConfig[T any](code string, required bool, features ...string) *core.AppConfigCode[T] {
	c := core.NewJsonConfig[T](code)
	if _, dup := appConfigMap[c.Code()]; dup {
		panic("appconfigcode: duplicate code " + c.Code())
	}
	appConfigMap[c.Code()] = &AppConfig{AppConfig: c, Required: required, Features: features}
	return c
}

var (
	// Common
	AppConfigAppSettings = NewAppConfig[models.AppConfigAppSettings]("APP_SETTINGS", false)

	// Wholesale
	// WHOLESALE_EMAIL: Reply-To and Cc of the booking email to wholesales. Read when an email is queued;
	// a missing or invalid row queues it without them (and logs an error).
	AppConfigWholesaleEmail = NewAppConfig[models.AppConfigWholesaleEmail]("WHOLESALE_EMAIL", true, "wholesale-email")

	// Email
	// EMAIL_CONTACT_CARD: the contact card image in email footers (admin page). Missing = the built-in card.
	AppConfigEmailContactCard = NewAppConfig[models.AppConfigEmailContactCard]("EMAIL_CONTACT_CARD", false, "wholesale-email")

	// AI
	// AI_SETTINGS: Claude API key (encrypted), model and on/off for AI tour prefill (admin page
	// "ตั้งค่า AI", PUT /ai-settings). Missing = AI off.
	AppConfigAISettings = NewAppConfig[models.AppConfigAISettings]("AI_SETTINGS", false, "ai-extract")
)
