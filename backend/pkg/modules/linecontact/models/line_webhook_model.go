package models

// LineWebhookRequest is the raw webhook body and its X-Line-Signature header.
type LineWebhookRequest struct {
	Body      []byte
	Signature string
}

// LineWebhookResponse is empty: LINE only needs a 200.
type LineWebhookResponse struct{}

// WebhookPayload is the subset of the Messaging API webhook body the backend reads.
type WebhookPayload struct {
	Destination string         `json:"destination"`
	Events      []WebhookEvent `json:"events"`
}

type WebhookEvent struct {
	Type      string        `json:"type"`      // follow, unfollow, message, ...
	Timestamp int64         `json:"timestamp"` // milliseconds since the epoch
	Source    WebhookSource `json:"source"`
}

type WebhookSource struct {
	Type   string `json:"type"` // user, group, room
	UserID string `json:"userId"`
}
