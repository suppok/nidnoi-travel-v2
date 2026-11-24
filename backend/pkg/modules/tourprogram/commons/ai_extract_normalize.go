package commons

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Limits shared with the create/update validation of tour programs and periods.
const (
	maxNameRunes      = 200
	maxHighlightRunes = 500
	maxDays           = 60
	maxPeriodDays     = 90
)

// AIWarning is shown to staff above the prefilled form. Field names the form item it is about
// ("name", "countryIds", "periods[3].startDate"), "" for the whole document.
type AIWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NormalizedTour is the extraction after Go post-processing, still by name (no database ids):
// Gregorian ISO dates, whole-baht prices, a deduplicated period list. Golden eval fixtures
// (backend/ex/*.expected.json) are this shape.
type NormalizedTour struct {
	Name              string             `json:"name"`
	Days              *int               `json:"days"`
	Nights            *int               `json:"nights"`
	Highlight         string             `json:"highlight"`
	AirlineCode       string             `json:"airlineCode"`
	AirlineName       string             `json:"airlineName"`
	WholesaleTourCode string             `json:"wholesaleTourCode"`
	PrintedTourCode   string             `json:"printedTourCode"`
	Countries         []NormalizedPlace  `json:"countries"`
	Cities            []NormalizedCity   `json:"cities"`
	Periods           []NormalizedPeriod `json:"periods"`
	DepositPerPerson  *int               `json:"depositPerPerson"`
	JoinLandPrice     *int               `json:"joinLandPrice"`
	MinGroupSize      *int               `json:"minGroupSize"`
	OtherCharges      []string           `json:"otherCharges"`
	Confidence        AIConfidence       `json:"confidence"`
}

type NormalizedPlace struct {
	NameTh string `json:"nameTh"`
	NameEn string `json:"nameEn"`
}

type NormalizedCity struct {
	NameTh        string `json:"nameTh"`
	NameEn        string `json:"nameEn"`
	CountryNameEn string `json:"countryNameEn"`
}

type NormalizedPeriod struct {
	StartDate         string `json:"startDate"` // YYYY-MM-DD
	EndDate           string `json:"endDate"`
	Price             *int   `json:"price"` // adult
	ChildWithBedPrice *int   `json:"childWithBedPrice"`
	ChildNoBedPrice   *int   `json:"childNoBedPrice"`
	InfantPrice       *int   `json:"infantPrice"`
	SingleSupplement  *int   `json:"singleSupplement"`
	Note              string `json:"note"`
}

// NormalizeAIExtraction checks and converts the model's answer. It never trusts it: unusable values
// become empty with a warning. "", 0 (days, year, minGroupSize) and -1 (nights) mean "not found". today (Thailand date) only matters when a period has no year at all.
func NormalizeAIExtraction(x *AIExtraction, today time.Time) (*NormalizedTour, []AIWarning) {
	var warns []AIWarning
	warn := func(field, format string, args ...any) {
		warns = append(warns, AIWarning{Field: field, Message: fmt.Sprintf(format, args...)})
	}

	out := &NormalizedTour{
		Name:              truncateRunes(clean(x.Name), maxNameRunes),
		Highlight:         truncateRunes(clean(x.Highlight), maxHighlightRunes),
		AirlineCode:       strings.ToUpper(clean(x.Airline.IATACode)),
		AirlineName:       clean(x.Airline.Name),
		WholesaleTourCode: truncateRunes(clean(x.WholesaleTour), 50),
		PrintedTourCode:   strings.ToUpper(clean(x.PrintedTourCode)),
		Countries:         []NormalizedPlace{},
		Cities:            []NormalizedCity{},
		Periods:           []NormalizedPeriod{},
		OtherCharges:      []string{},
		Confidence:        normalizeConfidence(x.Confidence),
	}
	if clean(x.Name) != "" && utf8.RuneCountInString(clean(x.Name)) > maxNameRunes {
		warn("name", "ชื่อโปรแกรมยาวเกิน %d ตัวอักษร ตัดให้แล้ว", maxNameRunes)
	}
	if out.Name == "" {
		warn("name", "ไม่พบชื่อโปรแกรมในไฟล์")
	}
	if out.AirlineCode != "" && !regexp.MustCompile(`^[A-Z0-9]{2,3}$`).MatchString(out.AirlineCode) {
		out.AirlineCode = ""
	}

	// Days / nights
	if x.Days != nil && *x.Days >= 1 && *x.Days <= maxDays {
		out.Days = intPtr(*x.Days)
	} else {
		warn("days", "ไม่พบจำนวนวัน")
	}
	if x.Nights != nil && out.Days != nil && *x.Nights >= 0 && *x.Nights <= *out.Days {
		out.Nights = intPtr(*x.Nights)
	} else {
		warn("nights", "ไม่พบจำนวนคืน")
	}

	// Taxonomy (names only here; matched to ids in ResolveAIDraft)
	seen := map[string]bool{}
	for _, c := range x.Countries {
		p := NormalizedPlace{NameTh: clean(c.NameTh), NameEn: clean(c.NameEn)}
		key := strings.ToLower(p.NameEn + "|" + p.NameTh)
		if (p.NameTh == "" && p.NameEn == "") || seen[key] {
			continue
		}
		seen[key] = true
		out.Countries = append(out.Countries, p)
	}
	if len(out.Countries) == 0 {
		warn("countryIds", "ไม่พบประเทศปลายทางในไฟล์")
	}
	for _, c := range x.Cities {
		p := NormalizedCity{NameTh: clean(c.NameTh), NameEn: clean(c.NameEn), CountryNameEn: clean(c.CountryNameEn)}
		key := "city|" + strings.ToLower(p.NameEn+"|"+p.NameTh)
		if (p.NameTh == "" && p.NameEn == "") || seen[key] {
			continue
		}
		seen[key] = true
		out.Cities = append(out.Cities, p)
	}

	// Extras
	// Prices never warn: a price that can't be read is left empty for staff.
	out.DepositPerPerson = parsePrice(x.Extras.DepositPerPerson)
	out.JoinLandPrice = parsePrice(x.Extras.JoinLandPrice)
	if x.Extras.MinGroupSize != nil && *x.Extras.MinGroupSize > 0 && *x.Extras.MinGroupSize < 1000 {
		out.MinGroupSize = intPtr(*x.Extras.MinGroupSize)
	}
	for _, line := range x.Extras.OtherCharges {
		if s := strings.TrimSpace(line); s != "" {
			out.OtherCharges = append(out.OtherCharges, truncateRunes(s, 300))
		}
	}

	// Periods
	type key struct{ start, end string }
	byDates := map[key]int{}
	for i, p := range x.Periods {
		field := fmt.Sprintf("periods[%d]", len(out.Periods))
		row := strings.TrimSpace(p.SourceText)
		if row == "" {
			row = fmt.Sprintf("รอบที่ %d", i+1)
		}
		start, end, ok := resolvePeriodDates(p.Start, p.End, today)
		if !ok {
			warn("", "ข้ามรอบ \"%s\": อ่านวันที่ไม่ได้", row)
			continue
		}
		if end.Before(start) {
			warn("", "ข้ามรอบ \"%s\": วันกลับอยู่ก่อนวันไป", row)
			continue
		}
		if end.Sub(start) > maxPeriodDays*24*time.Hour {
			warn("", "ข้ามรอบ \"%s\": ห่างกันเกิน %d วัน", row, maxPeriodDays)
			continue
		}
		np := NormalizedPeriod{StartDate: start.Format(time.DateOnly), EndDate: end.Format(time.DateOnly), Note: truncateRunes(clean(p.Label), 500)}
		if prev, dup := byDates[key{np.StartDate, np.EndDate}]; dup {
			warn(fmt.Sprintf("periods[%d]", prev), "รอบ %s – %s ซ้ำในไฟล์ ใช้ราคาจากแถวแรก", np.StartDate, np.EndDate)
			continue
		}
		if out.Days != nil {
			if span := int(end.Sub(start).Hours()/24) + 1; span != *out.Days {
				warn(field+".startDate", "รอบ %s – %s ยาว %d วัน ไม่ตรงกับโปรแกรม %d วัน", np.StartDate, np.EndDate, span, *out.Days)
			}
		}

		np.Price = parsePrice(p.AdultPrice)
		np.ChildWithBedPrice = parsePrice(p.ChildWithBed)
		np.ChildNoBedPrice = parsePrice(p.ChildNoBed)
		np.InfantPrice = parsePrice(p.Infant)
		if single := parsePrice(p.SingleRoom); single != nil {
			kind := p.SingleRoomKind
			if kind != "supplement" && kind != "total" && np.Price != nil {
				kind = "supplement"
				if *single >= *np.Price {
					kind = "total"
				}
			}
			switch {
			case kind == "supplement":
				np.SingleSupplement = single
			case kind == "total" && np.Price != nil && *single > *np.Price:
				np.SingleSupplement = intPtr(*single - *np.Price)
			}
		}
		byDates[key{np.StartDate, np.EndDate}] = len(out.Periods)
		out.Periods = append(out.Periods, np)
	}
	if len(out.Periods) == 0 {
		warn("periods", "ไม่พบรอบเดินทางในไฟล์")
	}
	if !today.IsZero() {
		past := 0
		for _, p := range out.Periods {
			if p.StartDate <= today.Format(time.DateOnly) {
				past++
			}
		}
		if past > 0 {
			warn("periods", "มี %d รอบที่ออกเดินทางไปแล้วหรือออกวันนี้ ตรวจสอบก่อนบันทึก", past)
		}
	}
	return out, warns
}

// resolvePeriodDates fills missing years and converts Buddhist-era years. A row often prints the
// year only on the end date ("28 ธ.ค. – 01 ม.ค. 2570"): the start takes the end's year, minus one
// when its month comes later in the year. With no year at all the next occurrence after today is used.
func resolvePeriodDates(s, e AIDate, today time.Time) (time.Time, time.Time, bool) {
	sy, sok := gregorianYear(s.Year)
	ey, eok := gregorianYear(e.Year)
	switch {
	case sok && !eok:
		ey = sy
		if e.Month < s.Month {
			ey++
		}
	case !sok && eok:
		sy = ey
		if s.Month > e.Month {
			sy--
		}
	case !sok && !eok:
		if today.IsZero() {
			return time.Time{}, time.Time{}, false
		}
		sy = today.Year()
		if s.Month < int(today.Month()) {
			sy++
		}
		ey = sy
		if e.Month < s.Month {
			ey++
		}
	}
	start, ok1 := validDate(sy, s.Month, s.Day)
	end, ok2 := validDate(ey, e.Month, e.Day)
	return start, end, ok1 && ok2
}

// gregorianYear: 2-digit years are Buddhist era (69 → 2569); years after 2400 are Buddhist era.
func gregorianYear(y *int) (int, bool) {
	if y == nil {
		return 0, false
	}
	year := *y
	switch {
	case year >= 0 && year < 100:
		year += 2500
		fallthrough
	case year > 2400:
		year -= 543
	}
	if year < 2000 || year > 2100 {
		return 0, false
	}
	return year, true
}

func validDate(y, m, d int) (time.Time, bool) {
	if m < 1 || m > 12 || d < 1 || d > 31 {
		return time.Time{}, false
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return t, t.Day() == d && int(t.Month()) == m // rejects 31/04
}

var priceNumber = regexp.MustCompile(`\d{1,3}(?:,\d{3})+|\d+`)

// parsePrice reads a printed price ("23,999 บาท", "57,900.-", "6000") as whole baht. Decimals are
// dropped. Zero or no number is nil.
func parsePrice(s *string) *int {
	if s == nil {
		return nil
	}
	m := priceNumber.FindString(*s)
	if m == "" {
		return nil
	}
	n, err := strconv.Atoi(strings.ReplaceAll(m, ",", ""))
	if err != nil || n <= 0 || n > 10_000_000 {
		return nil
	}
	return &n
}

func normalizeConfidence(c AIConfidence) AIConfidence {
	level := func(s string) string {
		switch s {
		case "high", "medium", "low":
			return s
		}
		return "low"
	}
	return AIConfidence{Main: level(c.Main), Taxonomy: level(c.Taxonomy), Periods: level(c.Periods), Prices: level(c.Prices), Extras: level(c.Extras)}
}

func clean(s *string) string {
	if s == nil {
		return ""
	}
	return strings.Join(strings.Fields(*s), " ")
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func intPtr(n int) *int { return &n }

// formatBaht is 23999 → "23,999".
func formatBaht(n int) string {
	s := strconv.Itoa(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}
