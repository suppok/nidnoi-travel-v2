// Package notification renders outgoing emails (subject, HTML and plain text) from embedded
// templates and assets. It never sends: callers queue the result in email_outbox.
package notification

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"strconv"
	"strings"
	texttemplate "text/template"

	"backend/pkg/domain"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

//go:embed assets/*
var assetFS embed.FS

var (
	wholesaleBookingHTML = htmltemplate.Must(htmltemplate.New("wholesale_booking.html.tmpl").
				ParseFS(templateFS, "templates/wholesale_booking.html.tmpl"))
	wholesaleBookingText = texttemplate.Must(texttemplate.New("wholesale_booking.txt.tmpl").
				ParseFS(templateFS, "templates/wholesale_booking.txt.tmpl"))
)

// Email is a rendered email. Its HTML may reference inline images as "cid:<id>" (see inline_assets.go);
// the outbox sender attaches them.
type Email struct {
	Subject string
	HTML    string
	Text    string
}

// WholesaleGreetingName is who the wholesale email greets ("เรียน …"). To be decided.
const WholesaleGreetingName = "XXX"

// WholesaleBooking is the data of the booking email to a wholesale.
type WholesaleBooking struct {
	Greeting      string   // "เรียน XXX"
	CustomerName  string   // as stored
	TourName      string   // the booking's snapshot
	DateRange     string   // "11 - 14 ตุลาคม 2569"
	Travellers    int      // infants included
	Summary       string   // BookingSummaryText
	SummaryLines  []string // Summary split into lines (HTML)
	ContactCardID string   // inline image of the footer
}

// NewWholesaleBooking builds the email data from a saved LINE booking (with its BookingSource and
// SalesUser), its tour (with Wholesale and Airline; nil when it can't be loaded) and the content id of
// the footer's contact card (ContactCardID; "" = the built-in card).
func NewWholesaleBooking(b *domain.LineBooking, tour *domain.TourProgram, contactCardID string) WholesaleBooking {
	if contactCardID == "" {
		contactCardID = DefaultContactCardID
	}
	summary := BookingSummaryText(b, tour)
	return WholesaleBooking{
		Greeting:      "เรียน " + WholesaleGreetingName,
		CustomerName:  strings.TrimSpace(b.CustomerName),
		TourName:      b.TourName,
		DateRange:     ThaiDateRange(b.StartDate, b.EndDate),
		Travellers:    b.Travellers(),
		Summary:       summary,
		SummaryLines:  strings.Split(summary, "\n"),
		ContactCardID: contactCardID,
	}
}

// WholesaleBookingSubject: "BOOKING <customer name> // <tour name> // <dates> // <n>PAX".
func WholesaleBookingSubject(d WholesaleBooking) string {
	subject := "BOOKING " + d.CustomerName + " // " + d.TourName + " // " + d.DateRange + " // " + strconv.Itoa(d.Travellers) + "PAX"
	// Headers are one line.
	return strings.Join(strings.Fields(subject), " ")
}

// RenderWholesaleBooking renders the booking email to a wholesale.
func RenderWholesaleBooking(d WholesaleBooking) (Email, error) {
	var html, text bytes.Buffer
	if err := wholesaleBookingHTML.Execute(&html, d); err != nil {
		return Email{}, err
	}
	if err := wholesaleBookingText.Execute(&text, d); err != nil {
		return Email{}, err
	}
	return Email{Subject: WholesaleBookingSubject(d), HTML: html.String(), Text: text.String()}, nil
}

// Baht formats whole baht with thousands separators: 12900 → "12,900".
func Baht(n int) string {
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
