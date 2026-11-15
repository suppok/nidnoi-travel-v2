package usecase

import (
	"context"
	"regexp"
	"time"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

const maxChatURLRunes = 500

// chatURLPattern: a chat in LINE OA Manager, as copied from the browser's address bar.
var chatURLPattern = regexp.MustCompile(`^https://chat\.line\.biz/\S+$`)

// saveChatURL stores the OA Manager chat URL sales pasted on the booking's LINE contact, so every
// booking with that chat can open it. An empty URL keeps the one already saved.
func (u *lineBookingUseCase) saveChatURL(ctx context.Context, appDB query.AppIDB, contact *domain.LineContact, url, userID string) error {
	if contact == nil || url == "" || url == contact.ChatURL {
		return nil
	}
	contact.ChatURL = url
	contact.UpdateDate = time.Now()
	contact.UpdateBy = userID
	_, err := u.repositoryRegistry.LineContactRepository.WithTx(appDB).Update(ctx, contact, "chat_url", "update_date", "update_by")
	return err
}
