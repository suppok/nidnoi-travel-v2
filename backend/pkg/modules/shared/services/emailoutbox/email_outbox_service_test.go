package emailoutbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/emailoutboxstate"
	"backend/pkg/core/config"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/client/mailer"
	"backend/pkg/modules/shared/services/notification"
	"backend/pkg/repository"
)

var now = time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)

func TestDecide(t *testing.T) {
	errX := errors.New("x")
	cases := []struct {
		name      string
		err       error
		retryable bool
		attempts  int
		state     int
		wait      time.Duration
	}{
		{"success", nil, false, 1, emailoutboxstate.Sent, 0},
		{"permanent", errX, false, 1, emailoutboxstate.Failed, 0},
		{"retry 1", errX, true, 1, emailoutboxstate.Pending, time.Minute},
		{"retry 2", errX, true, 2, emailoutboxstate.Pending, 5 * time.Minute},
		{"retry 3", errX, true, 3, emailoutboxstate.Pending, 30 * time.Minute},
		{"retry 4", errX, true, 4, emailoutboxstate.Pending, 2 * time.Hour},
		{"max reached", errX, true, 5, emailoutboxstate.Failed, 0},
	}
	for _, c := range cases {
		d := Decide(c.err, c.retryable, c.attempts, 5, now)
		if d.State != c.state || !d.NextAttemptAt.Equal(now.Add(c.wait)) {
			t.Errorf("%s: %+v", c.name, d)
		}
	}
	// The last delay repeats when maxAttempts allows more tries than delays.
	if d := Decide(errX, true, 7, 10, now); d.State != emailoutboxstate.Pending || !d.NextAttemptAt.Equal(now.Add(2*time.Hour)) {
		t.Errorf("beyond the schedule: %+v", d)
	}
}

func TestNewRow(t *testing.T) {
	addr := Addresses{From: "nidnoi travel <booking@example.com>", To: []string{"a@x.com"}, Cc: []string{"owner@example.com"}, ReplyTo: "owner@example.com"}
	row := NewRow(addr, "b1", emailoutboxstate.TypeWholesaleBooking, notification.Email{Subject: "s"}, "u1", now)
	if row.FromEmail != addr.From || row.ReplyTo != "owner@example.com" || len(row.CcEmails) != 1 || len(row.ToEmails) != 1 ||
		row.State != emailoutboxstate.Pending || row.SendRound != 1 || !row.NextAttemptAt.Equal(now) || row.ID == "" {
		t.Errorf("row %+v", row)
	}
	addr.Cc = nil
	if row := NewRow(addr, "b1", emailoutboxstate.TypeWholesaleBooking, notification.Email{}, "u1", now); row.CcEmails == nil || len(row.CcEmails) != 0 {
		t.Errorf("empty cc must be an empty array (the column is NOT NULL), got %#v", row.CcEmails)
	}
}

// fakeRepo records Finish calls.
type fakeRepo struct {
	repository.EmailOutboxRepository
	rows     []domain.EmailOutbox
	finished map[string]repository.EmailOutboxOutcome
}

func (f *fakeRepo) WithTx(query.AppIDB) repository.EmailOutboxRepository { return f }

func (f *fakeRepo) ClaimDue(_ context.Context, _, _ time.Time, limit int) ([]domain.EmailOutbox, error) {
	n := min(limit, len(f.rows))
	out := f.rows[:n]
	f.rows = f.rows[n:]
	for i := range out {
		out[i].Attempts++
	}
	return out, nil
}

func (f *fakeRepo) Finish(_ context.Context, id string, _ int, o repository.EmailOutboxOutcome) error {
	f.finished[id] = o
	return nil
}

type fakeMailer map[string]error

func (m fakeMailer) Send(_ context.Context, msg mailer.Message) (string, error) {
	if err := m[msg.To[0]]; err != nil {
		return "", err
	}
	return "msg-" + msg.IdempotencyKey, nil
}

func TestSenderTick(t *testing.T) {
	repo := &fakeRepo{finished: map[string]repository.EmailOutboxOutcome{}, rows: []domain.EmailOutbox{
		{ID: "ok", ToEmails: []string{"ok@x.com"}, SendRound: 1},
		{ID: "busy", ToEmails: []string{"busy@x.com"}, SendRound: 1},
		{ID: "bad", ToEmails: []string{"bad@x.com"}, SendRound: 2},
	}}
	m := fakeMailer{
		"busy@x.com": &mailer.SendError{Retryable: true, Status: 429, Message: "slow down"},
		"bad@x.com":  &mailer.SendError{Status: 422, Message: "invalid to"},
	}
	s := NewSender(repo, m, zap.NewNop(), config.MailOutboxConfig{BatchSize: 2, MaxAttempts: 5, TickInterval: time.Second})
	s.now = func() time.Time { return now }
	s.tick(context.Background())

	if o := repo.finished["ok"]; o.State != emailoutboxstate.Sent || o.ProviderMessageID != "msg-ok-1" || o.SentAt == nil {
		t.Errorf("ok: %+v", o)
	}
	if o := repo.finished["busy"]; o.State != emailoutboxstate.Pending || !o.NextAttemptAt.Equal(now.Add(time.Minute)) || o.LastError == "" {
		t.Errorf("busy: %+v", o)
	}
	if o := repo.finished["bad"]; o.State != emailoutboxstate.Failed || o.LastError != "mail: status 422: invalid to" {
		t.Errorf("bad: %+v", o)
	}
}
