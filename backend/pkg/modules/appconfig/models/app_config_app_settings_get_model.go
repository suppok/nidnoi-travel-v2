package models

import coremodels "backend/pkg/core/api/models"

type AppConfigAppSettingsGetRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AppConfigAppSettingsGetResponse struct {
	MaxFileSize int `json:"maxFileSize"` // MB
}
