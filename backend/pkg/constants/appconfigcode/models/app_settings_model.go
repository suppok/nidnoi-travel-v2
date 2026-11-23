package models

import "backend/pkg/core/validation"

// AppConfigAppSettings is the value of APP_SETTINGS.
type AppConfigAppSettings struct {
	MaxFileSize int `json:"maxFileSize"` // MB
}

func (m AppConfigAppSettings) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.MaxFileSize, validation.Required, validation.Min(1), validation.Max(100)),
	)
}
