package linecontact

import (
	"context"

	"backend/pkg/modules/linecontact/models"
)

type LineContactUseCase interface {
	// HandleWebhook verifies a Messaging API webhook and records follow, unfollow and message events.
	HandleWebhook(ctx context.Context, request *models.LineWebhookRequest) (*models.LineWebhookResponse, error)
	List(ctx context.Context, request *models.LineContactListRequest) (*models.LineContactListResponse, error)
}
