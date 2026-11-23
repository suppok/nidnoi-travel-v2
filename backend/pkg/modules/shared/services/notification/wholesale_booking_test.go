package notification

import (
	"strings"
	"testing"
	"time"

	"backend/pkg/domain"
)

func booking() *domain.LineBooking {
	return &domain.LineBooking{
		BookingNo:        "26X8K9AQ",
		TourCode:         "NIDN000123",
		TourName:         "ญี่ปุ่น <script>alert(1)</script> & โอซาก้า",
		StartDate:        time.Date(2026, 12, 5, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 12, 10, 0, 0, 0, 0, time.UTC),
		AdultTwinCount:   2,
		ChildNoBedCount:  1,
		InfantCount:      1,
		TotalPrice:       65900,
		CustomerName:     "สมชาย ใจดี",
		IsDeposit:        true,
		DepositPerPerson: 5000,
	}
}

func TestSubject(t *testing.T) {
	d := NewWholesaleBooking(booking(), "Zego", "แนน")
	d.TourName = "ญี่ปุ่น  โตเกียว\nฟูจิ"
	want := "[ยืนยันชำระเงิน] 26X8K9AQ – ญี่ปุ่น โตเกียว ฟูจิ 5/12/2569"
	if got := WholesaleBookingSubject(d); got != want {
		t.Errorf("subject = %q, want %q", got, want)
	}
}

func TestRenderWholesaleBooking(t *testing.T) {
	email, err := RenderWholesaleBooking(NewWholesaleBooking(booking(), "Zego Travel", "แนน"))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"26X8K9AQ", "NIDN000123", "5/12/2569 – 10/12/2569", "ผู้ใหญ่ (พักคู่)", "รวม", "4 ท่าน",
		"65,900 บาท", "ยอดมัดจำที่ชำระแล้ว", "20,000 บาท", "แนน", "รบกวนตอบกลับอีเมลนี้", "Zego Travel"} {
		if !strings.Contains(email.HTML, s) {
			t.Errorf("HTML missing %q", s)
		}
		if !strings.Contains(email.Text, s) {
			t.Errorf("text missing %q", s)
		}
	}
	if strings.Contains(email.HTML, "<script>") {
		t.Error("tour name not escaped in HTML")
	}
	if !strings.Contains(email.HTML, "&lt;script&gt;") {
		t.Error("escaped tour name missing from HTML")
	}
	if !strings.Contains(email.Text, "<script>alert(1)</script> & โอซาก้า") || strings.Contains(email.Text, "&lt;") {
		t.Error("text version must be plain, not HTML-escaped")
	}
	if strings.Contains(email.HTML, "เด็ก (มีเตียง)") {
		t.Error("rows with no travellers must be left out")
	}
	if strings.Contains(email.HTML, "<img") || strings.Contains(email.HTML, "http") {
		t.Error("no external images or links")
	}
}

func TestFullPayment(t *testing.T) {
	b := booking()
	b.IsDeposit = false
	d := NewWholesaleBooking(b, "", "")
	if d.PaidAmount != 65900 {
		t.Errorf("paid = %d", d.PaidAmount)
	}
	email, err := RenderWholesaleBooking(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(email.Text, "ยอดที่ชำระแล้ว (เต็มจำนวน): 65,900 บาท") || !strings.Contains(email.Text, "เรียน ฝ่ายจอง") {
		t.Errorf("text:\n%s", email.Text)
	}
}

func TestBaht(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 65900: "65,900", 1234567: "1,234,567", -1500: "-1,500"} {
		if got := Baht(n); got != want {
			t.Errorf("Baht(%d) = %q, want %q", n, got, want)
		}
	}
}
