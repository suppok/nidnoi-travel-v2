package models

import (
	"time"

	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

type AppConfigEmailContactCardGetRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// AppConfigEmailContactCardResponse is the contact card image at the bottom of emails.
type AppConfigEmailContactCardResponse struct {
	// ImageURL is the uploaded card; "" = the built-in card is used.
	ImageURL string `json:"imageUrl"`
	// DefaultImage is the built-in card as a data: URL, for the preview.
	DefaultImage string    `json:"defaultImage"`
	UpdateDate   null.Time `json:"updateDate" swaggertype:"string"` // null until an admin saves one
	UpdateBy     string    `json:"updateBy"`                        // name of who saved it; "" for none
}

// AppConfigEmailContactCardUpdateRequest: imageUrl from POST /media/upload (purpose
// email-contact-card), or "" to go back to the built-in card.
type AppConfigEmailContactCardUpdateRequest struct {
	ImageURL string `json:"imageUrl"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// UpdateDate helper for the response.
func NullTime(t time.Time) null.Time {
	if t.IsZero() {
		return null.Time{}
	}
	return null.TimeFrom(t)
}
