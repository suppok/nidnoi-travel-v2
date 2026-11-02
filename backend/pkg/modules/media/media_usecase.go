package media

import (
	"context"

	"backend/pkg/modules/media/models"
)

type MediaUseCase interface {
	Upload(ctx context.Context, request *models.MediaUploadRequest) (*models.MediaUploadResponse, error)
}
