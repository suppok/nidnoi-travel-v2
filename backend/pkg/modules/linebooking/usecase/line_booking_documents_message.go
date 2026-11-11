package usecase

import (
	"fmt"
	"time"

	"backend/pkg/domain"
)

// documentButtons are the PDFs of b in plan order, with their button labels (LINE allows 20
// characters per label).
func documentButtons(b *domain.LineBooking) [][2]string {
	if b.IsDeposit {
		return [][2]string{{"AI ใบรับเงินมัดจำ", b.AIFile}, {"IV ใบแจ้งหนี้", b.IVFile}}
	}
	return [][2]string{{"RE ใบเสร็จรับเงิน", b.REFile}}
}

// thaiDate is d/m/yyyy in the Buddhist era, as Thai customers read dates.
func thaiDate(t time.Time) string {
	return fmt.Sprintf("%d/%d/%d", t.Day(), int(t.Month()), t.Year()+543)
}

// documentsMessages is the LINE push for the customer: one Flex card with the booking and a button
// per PDF (a bot can't attach files, so each button opens the PDF in LINE's browser).
func documentsMessages(b *domain.LineBooking) []any {
	buttons := []any{}
	for i, d := range documentButtons(b) {
		style := "secondary"
		if i == 0 {
			style = "primary"
		}
		buttons = append(buttons, map[string]any{
			"type": "button", "style": style, "height": "sm",
			"action": map[string]any{"type": "uri", "label": d[0], "uri": d[1]},
		})
	}
	text := func(s, size, color string, bold bool) map[string]any {
		m := map[string]any{"type": "text", "text": s, "size": size, "wrap": true}
		if color != "" {
			m["color"] = color
		}
		if bold {
			m["weight"] = "bold"
		}
		return m
	}
	title := "ใบเสร็จรับเงิน"
	if b.IsDeposit {
		title = "เอกสารรับเงินมัดจำ"
	}
	return []any{map[string]any{
		"type":    "flex",
		"altText": fmt.Sprintf("%s ใบจอง %s", title, b.BookingNo),
		"contents": map[string]any{
			"type": "bubble",
			"body": map[string]any{
				"type": "box", "layout": "vertical", "spacing": "sm",
				"contents": []any{
					text(title, "lg", "", true),
					text("ใบจอง "+b.BookingNo, "sm", "#666666", false),
					text(b.TourCode+" "+b.TourName, "sm", "", false),
					text(fmt.Sprintf("เดินทาง %s – %s", thaiDate(b.StartDate), thaiDate(b.EndDate)), "sm", "#666666", false),
					text("กดปุ่มด้านล่างเพื่อเปิดเอกสาร (PDF)", "xs", "#999999", false),
				},
			},
			"footer": map[string]any{"type": "box", "layout": "vertical", "spacing": "sm", "contents": buttons},
		},
	}}
}
