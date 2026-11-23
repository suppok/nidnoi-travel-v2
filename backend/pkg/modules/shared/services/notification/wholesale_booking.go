// Package notification renders outgoing emails (subject, HTML and plain text) from embedded
// templates. It never sends: callers queue the result in email_outbox.
package notification

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"strconv"
	"strings"
	texttemplate "text/template"

	"backend/pkg/core/utils"
	"backend/pkg/domain"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var funcs = map[string]any{"baht": Baht}

var (
	wholesaleBookingHTML = htmltemplate.Must(htmltemplate.New("wholesale_booking.html.tmpl").Funcs(funcs).
				ParseFS(templateFS, "templates/wholesale_booking.html.tmpl"))
	wholesaleBookingText = texttemplate.Must(texttemplate.New("wholesale_booking.txt.tmpl").Funcs(funcs).
				ParseFS(templateFS, "templates/wholesale_booking.txt.tmpl"))
)

// Email is a rendered email.
type Email struct {
	Subject string
	HTML    string
	Text    string
}

// TravellerLine is one row of the travellers table.
type TravellerLine struct {
	Label string
	Count int
}

// WholesaleBooking is the data of the booking email to a wholesale.
type WholesaleBooking struct {
	WholesaleName   string
	BookingNo       string
	TourCode        string
	TourName        string
	StartDate       string // Thai d/m/yyyy
	EndDate         string
	CustomerName    string
	Travellers      []TravellerLine
	TotalTravellers int
	TotalPrice      int
	IsDeposit       bool
	PaidAmount      int // the deposit (with its card fee) or the full total
	SalesName       string
}

// NewWholesaleBooking builds the email data from a saved LINE booking.
func NewWholesaleBooking(b *domain.LineBooking, wholesaleName, salesName string) WholesaleBooking {
	rows := []TravellerLine{
		{"ผู้ใหญ่ (พักคู่)", b.AdultTwinCount},
		{"ผู้ใหญ่ (พักเดี่ยว)", b.AdultSingleCount},
		{"เด็ก (มีเตียง)", b.ChildWithBedCount},
		{"เด็ก (ไม่มีเตียง)", b.ChildNoBedCount},
		{"ทารก", b.InfantCount},
	}
	travellers := make([]TravellerLine, 0, len(rows))
	total := 0
	for _, r := range rows {
		if r.Count > 0 {
			travellers = append(travellers, r)
			total += r.Count
		}
	}
	paid := b.TotalPrice
	if b.IsDeposit {
		paid = b.DepositPerPerson*b.Travellers() + b.CardFee
	}
	return WholesaleBooking{
		WholesaleName:   wholesaleName,
		BookingNo:       b.BookingNo,
		TourCode:        b.TourCode,
		TourName:        b.TourName,
		StartDate:       utils.ThaiDate(b.StartDate),
		EndDate:         utils.ThaiDate(b.EndDate),
		CustomerName:    b.CustomerName,
		Travellers:      travellers,
		TotalTravellers: total,
		TotalPrice:      b.TotalPrice,
		IsDeposit:       b.IsDeposit,
		PaidAmount:      paid,
		SalesName:       salesName,
	}
}

// WholesaleBookingSubject: "[ยืนยันชำระเงิน] <booking no> – <tour name> <departure date>", so the owner
// can search Gmail by booking no.
func WholesaleBookingSubject(d WholesaleBooking) string {
	subject := "[ยืนยันชำระเงิน] " + d.BookingNo + " – " + d.TourName + " " + d.StartDate
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
