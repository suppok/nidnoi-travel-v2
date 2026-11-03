package site

import (
	"context"

	"backend/pkg/modules/site/models"
)

type SiteUseCase interface {
	Rebuild(ctx context.Context, request *models.SiteRebuildRequest) (*models.SiteRebuildResponse, error)
	BuildList(ctx context.Context, request *models.SiteBuildListRequest) (*models.SiteBuildListResponse, error)
}
