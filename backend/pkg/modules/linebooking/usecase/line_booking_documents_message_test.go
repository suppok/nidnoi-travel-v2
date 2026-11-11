package usecase

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"backend/pkg/domain"
)

func TestDocumentsMessages(t *testing.T) {
	start := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	b := &domain.LineBooking{
		BookingNo: "26X8K9AQ", TourCode: "NIDN000001", TourName: "ญี่ปุ่น", StartDate: start, EndDate: start.AddDate(0, 0, 4),
		IsDeposit: true, AIFile: "https://media.test/bookings/documents/a/ai.pdf", IVFile: "https://media.test/bookings/documents/b/iv.pdf",
	}
	raw, err := json.Marshal(documentsMessages(b))
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, want := range []string{`"type":"flex"`, b.AIFile, b.IVFile, "26X8K9AQ", "1/12/2569"} {
		if !strings.Contains(s, want) {
			t.Fatalf("message missing %q: %s", want, s)
		}
	}
	if strings.Contains(s, "RE ") {
		t.Fatalf("deposit message must not have an RE button: %s", s)
	}

	full := &domain.LineBooking{REFile: "https://media.test/bookings/documents/c/re.pdf"}
	if got := documentButtons(full); len(got) != 1 || got[0][1] != full.REFile {
		t.Fatalf("full payment buttons = %v", got)
	}
	// LINE rejects button labels over 20 characters.
	for _, d := range append(documentButtons(b), documentButtons(full)...) {
		if n := utf8.RuneCountInString(d[0]); n > 20 {
			t.Fatalf("label %q has %d characters", d[0], n)
		}
	}
}

func TestPushTarget(t *testing.T) {
	following := &domain.LineContact{ID: "c1", LineUserID: "U1", Following: true}
	blocked := &domain.LineContact{ID: "c2", LineUserID: "U2", Following: false}
	if pushTarget(&domain.LineBooking{LineContact: following}) != following {
		t.Fatal("a following contact is the target")
	}
	if pushTarget(&domain.LineBooking{LineContact: blocked}) != nil {
		t.Fatal("a blocked contact gets no push")
	}
	if pushTarget(&domain.LineBooking{}) != nil {
		t.Fatal("no contact, no push")
	}
}
