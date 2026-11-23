package models

import "backend/pkg/core/validation"

// AppConfigEmailContactCard is the value of EMAIL_CONTACT_CARD: the contact card image at the bottom of
// emails, uploaded on the backoffice page "รูปติดต่อท้ายอีเมล" (POST /media/upload, purpose
// email-contact-card). "" = the built-in card.
type AppConfigEmailContactCard struct {
	ImageURL string `json:"imageUrl"`
}

func (m AppConfigEmailContactCard) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.ImageURL, validation.RuneLength(0, 1000)),
	)
}
