package models

import coremodels "backend/pkg/core/api/models"

type SiteRebuildRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// SiteRebuildResponse confirms Cloudflare accepted the build. It takes a few minutes to go live.
type SiteRebuildResponse struct {
	Triggered bool `json:"triggered"`
}
