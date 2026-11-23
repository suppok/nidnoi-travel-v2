package notification

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/pkg/constants/linebookingpayment"
	"backend/pkg/domain"
)

// BookingSummaryText is the "สรุปยอดโอน" text of a saved LINE booking: the same sections and labels
// as the backoffice booking summary tool (frontend/backoffice/src/features/sales-tool/booking-summary/
// summary.ts buildSummaryText), built from what the booking stored. b needs its BookingSource and
// SalesUser; tour (may be nil) its Wholesale and Airline.
func BookingSummaryText(b *domain.LineBooking, tour *domain.TourProgram) string {
	var lines []string
	push := func(l ...string) { lines = append(lines, l...) }
	travellers := b.Travellers()

	source, shortName := "", ""
	if b.BookingSource != nil {
		source, shortName = b.BookingSource.Name, b.BookingSource.ShortName
	}
	refLabel := "ชื่อ"
	if shortName != "" {
		refLabel = "ชื่อ " + shortName
	}
	programme, pdf, airline, wholesale := b.TourName, "", "", ""
	if tour != nil {
		programme = fmt.Sprintf("%s (%d วัน %d คืน)", b.TourName, tour.Days, tour.Nights)
		pdf = tour.PDFFile.String
		if tour.Airline != nil {
			airline = tour.Airline.Name
		}
		if tour.Wholesale != nil {
			wholesale = tour.Wholesale.Name
		}
	}

	push("🔴สรุปรายละเอียดการจอง🔴")
	push("• จองผ่าน : " + orDash(source))
	push("• " + refLabel + " : " + orDash(b.RefName))
	push("• เลข booking : " + orDash(b.RefBookingNo))
	push("• ชื่อโปรแกรม : " + orDash(programme))
	push("• รหัสทัวร์ : " + orDash(b.TourCode))
	push("• โปรแกรมทัวร์ : " + orDash(pdf))
	push("• สายการบิน : " + orDash(airline))
	push("• Wholesale : " + orDash(wholesale))
	push("• วันเดินทาง : " + ThaiDateRange(b.StartDate, b.EndDate))
	push(fmt.Sprintf("• จำนวนผู้เดินทาง : %d ท่าน", travellers))

	adult := b.AdultTwinPrice
	line := "• ราคาทัวร์ผู้ใหญ่/ท่าน : " + money(adult)
	if b.AdultPriceBeforeDiscount.Valid && int(b.AdultPriceBeforeDiscount.Int64) > adult {
		line += " (จาก " + money(int(b.AdultPriceBeforeDiscount.Int64)) + ")"
	}
	push(line)
	priced := func(label string, count int, price int, unit string) {
		if count > 0 && price > 0 {
			push(fmt.Sprintf("• %s : %s (จำนวน %d %s)", label, money(price), count, unit))
		}
	}
	priced("ราคาเด็กอายุต่ำกว่า 20 ปี/ท่าน", b.ChildWithBedCount, int(b.ChildWithBedPrice.Int64), "ท่าน")
	priced("ราคาเด็กอายุต่ำกว่า 20 ปี (ไม่เสริมเตียง)/ท่าน", b.ChildNoBedCount, int(b.ChildNoBedPrice.Int64), "ท่าน")
	priced("ราคาเด็กทารก/ท่าน", b.InfantCount, int(b.InfantPrice.Int64), "ท่าน")

	charge := func(kind string) (int, int) {
		for _, c := range b.Charges {
			if c.Kind == kind {
				return c.Count, c.Price
			}
		}
		return 0, 0
	}
	count, price := charge(linebookingpayment.ChargeVisa)
	priced("ค่า visa/ท่าน", count, price, "ท่าน")
	count, price = charge(linebookingpayment.ChargeForeignPassport)
	priced("ค่า passport ต่างชาติ/ท่าน", count, price, "ท่าน")
	if b.AdultSinglePrice.Valid {
		priced("ค่าห้องพักเดี่ยว/ห้อง", b.AdultSingleCount, int(b.AdultSinglePrice.Int64)-adult, "ห้อง")
	}
	count, price = charge(linebookingpayment.ChargeJoinLand)
	priced("ค่า join land/ท่าน", count, price, "ท่าน")
	count, price = charge(linebookingpayment.ChargeFuelTax)
	priced("ค่าภาษีน้ำมัน/ท่าน", count, price, "ท่าน")
	count, price = charge(linebookingpayment.ChargeAirportTax)
	priced("ค่าภาษีสนามบิน/ท่าน", count, price, "ท่าน")
	for _, c := range b.Charges {
		if c.Kind != linebookingpayment.ChargeOption {
			continue
		}
		name := strings.TrimSpace(c.Name)
		if name == "" {
			name = "Option"
		}
		priced(name+"/ท่าน", c.Count, c.Price, "ท่าน")
	}
	if b.DiscountPerPerson > 0 {
		push("• ส่วนลด/ท่าน : " + money(b.DiscountPerPerson))
	}
	if b.PaymentMethod == linebookingpayment.Credit {
		push(fmt.Sprintf("• ค่าธรรมเนียมบัตรเครดิต (%s%%) : %s", percent(b.CreditFeeBps), money(b.CardFee)))
	}
	push("• รวมเป็นเงิน : " + money(b.TotalPrice))

	const seatCut = "(ขอสงวนสิทธิ์ตัดที่นั่งอัตโนมัติ หากไม่ชำระเงินภายในเวลาที่กำหนด)"
	push("", "🔴รายละเอียดการชำระเงิน🔴")
	if b.IsDeposit {
		deposit := b.DepositPerPerson * travellers
		remaining := b.TotalPrice - b.CardFee - deposit
		push(fmt.Sprintf("• กรุณาชำระเงินมัดจำยอดเงิน %s (ท่านละ %s) %s %s",
			money(deposit+b.CardFee), money(b.DepositPerPerson), deadline(b.DepositDueDate, b.DepositDueTime), seatCut))
		push(fmt.Sprintf("• กรุณาชำระเงินส่วนเหลือยอดเงิน %s ภายในวันที่ %s", money(remaining), thaiDatePtr(b.RemainingDueDate)))
	} else {
		push(fmt.Sprintf("• กรุณาชำระเงินเต็มจำนวน %s %s %s", money(b.TotalPrice), deadline(b.FullDueDate, b.FullDueTime), seatCut))
	}

	customer := "-"
	if name := strings.TrimSpace(b.CustomerName); name != "" {
		customer = "คุณ" + name
	}
	sales := "-"
	if b.SalesUser != nil {
		sales = b.SalesUser.Name
		if b.SalesUser.Phone != "" {
			sales += " " + FormatPhone(b.SalesUser.Phone)
		}
	}
	push(
		"",
		"✅ ราคาที่ท่านซื้อ ถือเป็นราคาที่ต่ำ ที่ท่านพึงพอใจและตัดสินใจดีแล้ว เมื่อมีที่หลุดเพิ่มเติมในกรณีต่าง ๆ อาจเป็นราคาที่ต่ำกว่าที่ท่านซื้อก็เป็นไปได้",
		"✅ ไม่สามารถคืนเงิน เปลี่ยนแปลง หรือยกเลิกได้ในทุกกรณี",
		"-------------------------------",
		"",
		"🔴ข้อมูลลูกค้า🔴",
		"• ผู้จอง : "+customer,
		"• เบอร์โทร : "+orDash(b.CustomerPhone),
		"• ที่อยู่ : "+orDash(b.CustomerAddress),
		"• ผู้ดูแล : "+sales,
	)
	return strings.Join(lines, "\n")
}

var thaiMonths = [...]string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}

// ThaiLongDate: "11 ตุลาคม 2569" (Buddhist era).
func ThaiLongDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), thaiMonths[t.Month()], t.Year()+543)
}

// ThaiDateRange: "11 - 14 ตุลาคม 2569", "28 ตุลาคม - 2 พฤศจิกายน 2569", or both dates in full across years.
func ThaiDateRange(start, end time.Time) string {
	switch {
	case start.Year() == end.Year() && start.Month() == end.Month():
		if start.Day() == end.Day() {
			return ThaiLongDate(start)
		}
		return fmt.Sprintf("%d - %s", start.Day(), ThaiLongDate(end))
	case start.Year() == end.Year():
		return fmt.Sprintf("%d %s - %s", start.Day(), thaiMonths[start.Month()], ThaiLongDate(end))
	default:
		return ThaiLongDate(start) + " - " + ThaiLongDate(end)
	}
}

var (
	phone10 = regexp.MustCompile(`^0\d{9}$`)
	phone9  = regexp.MustCompile(`^0\d{8}$`)
)

// FormatPhone mirrors the backoffice formatPhone: 0945466166 → "094 546 6166", 021234567 → "02 123 4567".
func FormatPhone(p string) string {
	switch {
	case phone10.MatchString(p):
		return p[:3] + " " + p[3:6] + " " + p[6:]
	case phone9.MatchString(p):
		return p[:2] + " " + p[2:5] + " " + p[5:]
	}
	return p
}

func money(v int) string { return Baht(v) + " บาท" }

func orDash(s string) string {
	if s = strings.TrimSpace(s); s == "" {
		return "-"
	}
	return s
}

func thaiDatePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return ThaiLongDate(*t)
}

func deadline(date *time.Time, clock string) string {
	s := "ภายในวันที่ " + thaiDatePtr(date)
	if clock != "" {
		s += " (ก่อน " + clock + " น.)"
	}
	return s
}

// percent formats basis points as a percentage without trailing zeros: 260 → "2.6", 300 → "3".
func percent(bps int) string {
	return strconv.FormatFloat(float64(bps)/100, 'f', -1, 64)
}
