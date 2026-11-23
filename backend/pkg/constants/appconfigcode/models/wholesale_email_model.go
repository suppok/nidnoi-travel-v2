package models

import "backend/pkg/core/validation"

const wholesaleEmailMaxCc = 10

// AppConfigWholesaleEmail is the value of WHOLESALE_EMAIL: the Reply-To and Cc of every email sent to
// a wholesale (the owner's inbox, so replies and a copy land there). Empty = header not set.
type AppConfigWholesaleEmail struct {
	ReplyTo string   `json:"replyTo"`
	Cc      []string `json:"cc"`
}

func (m AppConfigWholesaleEmail) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.ReplyTo, validation.Email),
		validation.Field(&m.Cc,
			validation.Length(0, wholesaleEmailMaxCc),
			validation.Each(validation.Required, validation.Email),
		),
	)
}
