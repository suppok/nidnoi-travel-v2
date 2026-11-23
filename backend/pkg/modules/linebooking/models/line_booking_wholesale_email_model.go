package models

import (
	"time"

	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

// Results of queuing the wholesale email when a booking is saved on the entry page.
const (
	WholesaleEmailQueued   = "queued"   // queued; the sender delivers it within seconds
	WholesaleEmailDisabled = "disabled" // the wholesale has auto email off
	WholesaleEmailNoEmail  = "no_email" // auto email is on but the wholesale has no email address
)

// LineBookingWholesaleEmailResult tells the entry page what happened to the wholesale email.
type LineBookingWholesaleEmailResult struct {
	Result        string   `json:"result" enums:"queued,disabled,no_email"`
	WholesaleName string   `json:"wholesaleName"`
	To            []string `json:"to"` // the addresses it was queued to; empty unless queued
}

type LineBookingWholesaleEmailGetRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// LineBookingWholesaleEmailItem is the state of a booking's email to its wholesale.
type LineBookingWholesaleEmailItem struct {
	// State: 10 pending (queued or waiting to retry), 20 sent, 30 failed.
	State         int       `json:"state" enums:"10,20,30"`
	ToEmails      []string  `json:"toEmails"`
	CcEmails      []string  `json:"ccEmails"`
	ReplyTo       string    `json:"replyTo"`
	Subject       string    `json:"subject"`
	Attempts      int       `json:"attempts"`  // in the current round
	SendRound     int       `json:"sendRound"` // 1 + manual resends
	NextAttemptAt time.Time `json:"nextAttemptAt"`
	LastError     string    `json:"lastError"`
	SentAt        null.Time `json:"sentAt" swaggertype:"string"`
	CreateDate    time.Time `json:"createDate"`
	UpdateDate    time.Time `json:"updateDate"`
}

// LineBookingWholesaleEmailGetResponse: email is null when none was queued (auto email off or no
// address when the booking was saved, or a booking from before the feature).
type LineBookingWholesaleEmailGetResponse struct {
	Email *LineBookingWholesaleEmailItem `json:"email"`
}

type LineBookingWholesaleEmailResendRequest struct {
	ID string `json:"id"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingWholesaleEmailResendResponse struct {
	Email LineBookingWholesaleEmailItem `json:"email"`
}

func NewLineBookingWholesaleEmailItem(e *domain.EmailOutbox) LineBookingWholesaleEmailItem {
	return LineBookingWholesaleEmailItem{
		State:         e.State,
		ToEmails:      e.ToEmails,
		CcEmails:      e.CcEmails,
		ReplyTo:       e.ReplyTo,
		Subject:       e.Subject,
		Attempts:      e.Attempts,
		SendRound:     e.SendRound,
		NextAttemptAt: e.NextAttemptAt,
		LastError:     e.LastError,
		SentAt:        e.SentAt,
		CreateDate:    e.CreateDate,
		UpdateDate:    e.UpdateDate,
	}
}
