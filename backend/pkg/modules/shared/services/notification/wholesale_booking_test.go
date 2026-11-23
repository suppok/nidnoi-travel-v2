package notification

import (
	"strings"
	"testing"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/linebookingpayment"
	"backend/pkg/domain"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func booking() *domain.LineBooking {
	due := date(2026, 10, 9)
	return &domain.LineBooking{
		BookingNo:       "26X8K9AQ",
		TourCode:        "NIDN260865",
		TourName:        "ทัวร์เวียดนาม VIETNAM เวียดนามกลาง พักบานาฮิลล์",
		StartDate:       date(2026, 10, 11),
		EndDate:         date(2026, 10, 14),
		AdultTwinCount:  3,
		AdultTwinPrice:  11999,
		TotalPrice:      43497,
		CustomerName:    "มัฏฐณัฐ ภักดีสุวรรณ์",
		CustomerPhone:   "0811359151",
		CustomerAddress: "300/108 กรุงเทพมหานคร 10510",
		RefName:         "Nattie",
		PaymentMethod:   linebookingpayment.Transfer,
		FullDueDate:     &due,
		FullDueTime:     "13:40",
		Charges: []domain.LineBookingCharge{
			{Kind: linebookingpayment.ChargeForeignPassport, Count: 2, Price: 3000},
			{Kind: linebookingpayment.ChargeOption, Name: "FAMILY BUNG เตียง 2 ชั้น", Count: 1, Price: 1500},
		},
		BookingSource: &domain.BookingSource{Name: "LINE@ชมพู", ShortName: "LINE"},
		SalesUser:     &domain.AppUser{Name: "แอดมินกุ้ง", Phone: "0945466166"},
	}
}

func tour() *domain.TourProgram {
	return &domain.TourProgram{
		Days:      4,
		Nights:    3,
		PDFFile:   null.StringFrom("https://media.example.com/tours/documents/x/nidn260865.pdf"),
		Airline:   &domain.Airline{Name: "VIETJET AIR"},
		Wholesale: &domain.Wholesale{Name: "TOURNUMDEE"},
	}
}

func TestSubject(t *testing.T) {
	d := NewWholesaleBooking(booking(), tour(), "")
	want := "BOOKING มัฏฐณัฐ ภักดีสุวรรณ์ // ทัวร์เวียดนาม VIETNAM เวียดนามกลาง พักบานาฮิลล์ // 11 - 14 ตุลาคม 2569 // 3PAX"
	if got := WholesaleBookingSubject(d); got != want {
		t.Errorf("subject = %q\nwant      %q", got, want)
	}
}

func TestThaiDateRange(t *testing.T) {
	cases := map[[2]time.Time]string{
		{date(2026, 10, 11), date(2026, 10, 14)}: "11 - 14 ตุลาคม 2569",
		{date(2026, 10, 28), date(2026, 11, 2)}:  "28 ตุลาคม - 2 พฤศจิกายน 2569",
		{date(2026, 12, 30), date(2027, 1, 3)}:   "30 ธันวาคม 2569 - 3 มกราคม 2570",
		{date(2026, 5, 1), date(2026, 5, 1)}:     "1 พฤษภาคม 2569",
	}
	for in, want := range cases {
		if got := ThaiDateRange(in[0], in[1]); got != want {
			t.Errorf("ThaiDateRange(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestBookingSummaryText(t *testing.T) {
	text := BookingSummaryText(booking(), tour())
	for _, s := range []string{
		"🔴สรุปรายละเอียดการจอง🔴",
		"• จองผ่าน : LINE@ชมพู",
		"• ชื่อ LINE : Nattie",
		"• เลข booking : -",
		"• ชื่อโปรแกรม : ทัวร์เวียดนาม VIETNAM เวียดนามกลาง พักบานาฮิลล์ (4 วัน 3 คืน)",
		"• รหัสทัวร์ : NIDN260865",
		"• โปรแกรมทัวร์ : https://media.example.com/tours/documents/x/nidn260865.pdf",
		"• สายการบิน : VIETJET AIR",
		"• Wholesale : TOURNUMDEE",
		"• วันเดินทาง : 11 - 14 ตุลาคม 2569",
		"• จำนวนผู้เดินทาง : 3 ท่าน",
		"• ราคาทัวร์ผู้ใหญ่/ท่าน : 11,999 บาท",
		"• ค่า passport ต่างชาติ/ท่าน : 3,000 บาท (จำนวน 2 ท่าน)",
		"• FAMILY BUNG เตียง 2 ชั้น/ท่าน : 1,500 บาท (จำนวน 1 ท่าน)",
		"• รวมเป็นเงิน : 43,497 บาท",
		"• กรุณาชำระเงินเต็มจำนวน 43,497 บาท ภายในวันที่ 9 ตุลาคม 2569 (ก่อน 13:40 น.)",
		"✅ ไม่สามารถคืนเงิน เปลี่ยนแปลง หรือยกเลิกได้ในทุกกรณี",
		"• ผู้จอง : คุณมัฏฐณัฐ ภักดีสุวรรณ์",
		"• ผู้ดูแล : แอดมินกุ้ง 094 546 6166",
	} {
		if !strings.Contains(text, s) {
			t.Errorf("summary missing %q\n%s", s, text)
		}
	}
	if strings.Contains(text, "ค่าธรรมเนียมบัตรเครดิต") || strings.Contains(text, "ราคาเด็ก") {
		t.Error("unused lines must be left out")
	}
}

func TestDepositAndCreditSummary(t *testing.T) {
	b := booking()
	b.IsDeposit = true
	b.DepositPerPerson = 5000
	b.PaymentMethod = linebookingpayment.Credit
	b.CreditFeeBps = 260
	b.CardFee = 390 // 2.6% of the 15,000 deposit
	b.TotalPrice = 43497 + 390
	remaining := date(2026, 9, 11)
	b.RemainingDueDate = &remaining
	b.DepositDueDate = b.FullDueDate
	text := BookingSummaryText(b, nil)
	for _, s := range []string{
		"• ค่าธรรมเนียมบัตรเครดิต (2.6%) : 390 บาท",
		"• กรุณาชำระเงินมัดจำยอดเงิน 15,390 บาท (ท่านละ 5,000 บาท) ภายในวันที่ 9 ตุลาคม 2569",
		"• กรุณาชำระเงินส่วนเหลือยอดเงิน 28,497 บาท ภายในวันที่ 11 กันยายน 2569",
		"• Wholesale : -", // no tour loaded
	} {
		if !strings.Contains(text, s) {
			t.Errorf("summary missing %q\n%s", s, text)
		}
	}
}

func TestRenderWholesaleBooking(t *testing.T) {
	b := booking()
	b.CustomerName = "<script>alert(1)</script>"
	email, err := RenderWholesaleBooking(NewWholesaleBooking(b, tour(), ""))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"เรียน XXX", "🔴สรุปรายละเอียดการจอง🔴", "ห้ามเมลมาแจ้งอย่างเดียวนะคะ", "ขอบคุณมากค่ะ", "Nidnoi Travel", "02-044-6161 , 02-073-6161"} {
		if !strings.Contains(email.HTML, s) || !strings.Contains(email.Text, s) {
			t.Errorf("missing %q", s)
		}
	}
	if !strings.HasPrefix(email.Text, "เรียน XXX\n\n🔴สรุปรายละเอียดการจอง🔴") {
		t.Errorf("text must start with the greeting and the summary:\n%s", email.Text)
	}
	if !strings.Contains(email.HTML, `src="cid:contact-card.png"`) {
		t.Error("HTML must show the contact card inline")
	}
	if strings.Contains(email.HTML, "<script>") || !strings.Contains(email.HTML, "&lt;script&gt;") {
		t.Error("data must be escaped in HTML")
	}
	if !strings.Contains(email.Text, "<script>alert(1)</script>") {
		t.Error("text version must be plain, not HTML-escaped")
	}
	if !strings.Contains(email.HTML, "<br>• จองผ่าน : LINE@ชมพู") {
		t.Error("summary lines must be separated by <br>")
	}
}

func TestContactCard(t *testing.T) {
	const base = "https://media.example.com"
	email, err := RenderWholesaleBooking(NewWholesaleBooking(booking(), tour(), "contact-card-0123456789abcdef01234567.png"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(email.HTML, `src="cid:contact-card-0123456789abcdef01234567.png"`) {
		t.Error("HTML must reference the uploaded card")
	}
	if ids := ContentIDs(email.HTML + `<img src="cid:contact-card-0123456789abcdef01234567.png">`); len(ids) != 1 {
		t.Errorf("ContentIDs = %v", ids)
	}

	cases := []struct {
		url, id string
		ok      bool
	}{
		{base + "/email/contact-cards/0123456789abcdef01234567.png", "contact-card-0123456789abcdef01234567.png", true},
		{base + "/email/contact-cards/0123456789abcdef01234567.jpg", "contact-card-0123456789abcdef01234567.jpg", true},
		{"", DefaultContactCardID, false},
		{"https://evil.example.com/email/contact-cards/0123456789abcdef01234567.png", DefaultContactCardID, false},
		{base + "/tours/cards/0123456789abcdef01234567.png", DefaultContactCardID, false},
		{base + "/email/contact-cards/../x.png", DefaultContactCardID, false},
		{base + "/email/contact-cards/0123456789abcdef01234567.webp", DefaultContactCardID, false},
	}
	for _, c := range cases {
		if id, ok := ContactCardID(c.url, base+"/"); id != c.id || ok != c.ok {
			t.Errorf("ContactCardID(%q) = %q, %v", c.url, id, ok)
		}
	}
	if k := UploadedCardKey("contact-card-0123456789abcdef01234567.png"); k != "email/contact-cards/0123456789abcdef01234567.png" {
		t.Errorf("UploadedCardKey = %q", k)
	}
	for _, id := range []string{DefaultContactCardID, "contact-card-../../x.png", "other.png"} {
		if k := UploadedCardKey(id); k != "" {
			t.Errorf("UploadedCardKey(%q) = %q, want empty", id, k)
		}
	}
	if card := DefaultContactCard(); len(card) < 1000 || string(card[1:4]) != "PNG" {
		t.Error("built-in card is not a PNG")
	}
	if ImageContentType("a.jpg") != "image/jpeg" || ImageContentType("a.png") != "image/png" {
		t.Error("ImageContentType")
	}
}

func TestFormatPhoneAndBaht(t *testing.T) {
	for in, want := range map[string]string{"0945466166": "094 546 6166", "020446161": "02 044 6161", "+66812345678": "+66812345678"} {
		if got := FormatPhone(in); got != want {
			t.Errorf("FormatPhone(%q) = %q, want %q", in, got, want)
		}
	}
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 43497: "43,497", 1234567: "1,234,567", -1500: "-1,500"} {
		if got := Baht(n); got != want {
			t.Errorf("Baht(%d) = %q, want %q", n, got, want)
		}
	}
}
