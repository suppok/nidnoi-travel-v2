package models

import (
	"time"

	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

type LineContactListRequest struct {
	Keyword   string `json:"keyword"`   // display name
	Following *bool  `json:"following"` // omitted = all
	Page      int    `json:"page"`      // default 1
	PageSize  int    `json:"pageSize"`  // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineContactListResponse struct {
	Items    []LineContactItem `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// LineContactItem is a LINE OA contact as shown in the staff contact picker.
type LineContactItem struct {
	ID            string      `json:"id"`
	LineUserID    string      `json:"lineUserId"`
	DisplayName   string      `json:"displayName"`
	PictureURL    null.String `json:"pictureUrl" swaggertype:"string"`
	Following     bool        `json:"following"`
	LastMessageAt *time.Time  `json:"lastMessageAt"`
}

func NewLineContactItem(c *domain.LineContact) LineContactItem {
	return LineContactItem{
		ID:            c.ID,
		LineUserID:    c.LineUserID,
		DisplayName:   c.DisplayName,
		PictureURL:    c.PictureURL,
		Following:     c.Following,
		LastMessageAt: c.LastMessageAt.Ptr(),
	}
}
