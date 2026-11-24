package commons

import (
	"fmt"
	"strings"
	"unicode"
)

// AIDraft prefills the create form: the shape of the tour program create DTO (TourProgramFields)
// plus the periods to create after it. It never carries a code (the backend generates it) and never
// creates taxonomy or airlines: unmatched names are warnings. The wholesale is always left for staff.
type AIDraft struct {
	RefID             *int            `json:"refId"` // always null
	Name              string          `json:"name"`
	WholesaleID       *string         `json:"wholesaleId"` // always null: staff pick it
	WholesaleTourCode string          `json:"wholesaleTourCode"`
	State             int             `json:"state"` // always draft (10)
	Days              *int            `json:"days"`
	Nights            *int            `json:"nights"`
	AirlineID         *string         `json:"airlineId"`
	Highlight         string          `json:"highlight"`
	Description       string          `json:"description"` // TBD: left empty for now
	BannerImage       *string         `json:"bannerImage"`
	CardImage         *string         `json:"cardImage"`
	PDFFile           *string         `json:"pdfFile"`
	DOCXFile          *string         `json:"docxFile"`
	InternalNote      string          `json:"internalNote"`
	CountryIDs        []string        `json:"countryIds"`
	CityIDs           []string        `json:"cityIds"`
	Images            []string        `json:"images"`
	Periods           []AIDraftPeriod `json:"periods"`
}

// AIDraftPeriod is a period create body without tourProgramId. Seats are never in the PDFs and stay
// null for staff to fill (no warning).
type AIDraftPeriod struct {
	StartDate           string `json:"startDate"`
	EndDate             string `json:"endDate"`
	Price               *int   `json:"price"`
	PriceBeforeDiscount *int   `json:"priceBeforeDiscount"`
	ChildWithBedPrice   *int   `json:"childWithBedPrice"`
	ChildNoBedPrice     *int   `json:"childNoBedPrice"`
	InfantPrice         *int   `json:"infantPrice"`
	SingleSupplement    *int   `json:"singleSupplement"`
	Deposit             *int   `json:"deposit"`
	SeatsTotal          *int   `json:"seatsTotal"`
	SeatsAvailable      *int   `json:"seatsAvailable"`
	Note                string `json:"note"`
}

// DraftStateDraft mirrors tourprogramstate draft; AI drafts never start as selling.
const DraftStateDraft = 10

// AILookup is the active master data names are matched against.
type AILookup struct {
	Countries []LookupCountry
	Cities    []LookupCity
	Airlines  []LookupCode
	// ExistingTourCode reports whether a tour with this code (case-insensitive) already exists.
	ExistingTourCode func(code string) bool
}

type LookupCountry struct{ ID, Slug, NameTh, NameEn string }
type LookupCity struct{ ID, CountryID, Slug, NameTh, NameEn string }
type LookupCode struct{ ID, Code, Name string }

// ResolveAIDraft matches names to ids and builds the draft. Matching is on slug, Thai or English
// name, ignoring case, spaces and punctuation.
func ResolveAIDraft(t *NormalizedTour, lookup AILookup) (*AIDraft, []AIWarning) {
	var warns []AIWarning
	warn := func(field, format string, args ...any) {
		warns = append(warns, AIWarning{Field: field, Message: fmt.Sprintf(format, args...)})
	}

	d := &AIDraft{
		Name:              t.Name,
		WholesaleTourCode: t.WholesaleTourCode,
		State:             DraftStateDraft,
		Days:              t.Days,
		Nights:            t.Nights,
		Highlight:         t.Highlight,
		CountryIDs:        []string{},
		CityIDs:           []string{},
		Images:            []string{},
		Periods:           []AIDraftPeriod{},
	}

	// Countries
	countryByKey := map[string]LookupCountry{}
	for _, c := range lookup.Countries {
		for _, k := range []string{c.Slug, c.NameTh, c.NameEn} {
			if key := matchKey(k); key != "" {
				countryByKey[key] = c
			}
		}
	}
	selected := map[string]bool{}
	for _, c := range t.Countries {
		found, ok := countryByKey[matchKey(c.NameEn)]
		if !ok {
			found, ok = countryByKey[matchKey(c.NameTh)]
		}
		if !ok {
			warn("countryIds", "ไม่พบประเทศ \"%s\" ในระบบ กรุณาเลือกเอง", placeName(c.NameTh, c.NameEn))
			continue
		}
		if !selected[found.ID] {
			selected[found.ID] = true
			d.CountryIDs = append(d.CountryIDs, found.ID)
		}
	}

	// Cities: each must belong to a selected country.
	cityByKey := map[string][]LookupCity{}
	for _, c := range lookup.Cities {
		for _, k := range []string{c.Slug, c.NameTh, c.NameEn} {
			if key := matchKey(k); key != "" {
				cityByKey[key] = append(cityByKey[key], c)
			}
		}
	}
	countryName := map[string]string{}
	for _, c := range lookup.Countries {
		countryName[c.ID] = c.NameTh
	}
	pickedCity := map[string]bool{}
	for _, c := range t.Cities {
		candidates := cityByKey[matchKey(c.NameEn)]
		if len(candidates) == 0 {
			candidates = cityByKey[matchKey(c.NameTh)]
		}
		if len(candidates) == 0 {
			warn("cityIds", "ไม่พบเมือง \"%s\" ในระบบ (ไม่บังคับ — เพิ่มเองได้)", placeName(c.NameTh, c.NameEn))
			continue
		}
		var match *LookupCity
		for i := range candidates {
			if selected[candidates[i].CountryID] {
				match = &candidates[i]
				break
			}
		}
		if match == nil {
			warn("cityIds", "เมือง \"%s\" อยู่ในประเทศ%s ที่ไม่ได้เลือกไว้ จึงไม่ได้เลือกเมืองนี้", placeName(c.NameTh, c.NameEn), countryName[candidates[0].CountryID])
			continue
		}
		if !pickedCity[match.ID] {
			pickedCity[match.ID] = true
			d.CityIDs = append(d.CityIDs, match.ID)
		}
	}

	// Airline: IATA code first, then name.
	if t.AirlineCode != "" || t.AirlineName != "" {
		if a, ok := findCode(lookup.Airlines, t.AirlineCode, t.AirlineName); ok {
			d.AirlineID = &a.ID
		} else {
			warn("airlineId", "ไม่พบสายการบิน \"%s\" ในระบบ กรุณาเลือกเอง", strings.TrimSpace(t.AirlineName+" "+t.AirlineCode))
		}
	}

	if t.PrintedTourCode != "" && lookup.ExistingTourCode != nil && lookup.ExistingTourCode(t.PrintedTourCode) {
		warn("", "ไฟล์นี้มีรหัสทัวร์ %s ซึ่งมีอยู่ในระบบแล้ว — อาจเป็นโปรแกรมซ้ำ (โปรแกรมใหม่จะได้รหัสใหม่)", t.PrintedTourCode)
	}

	// Periods
	for _, p := range t.Periods {
		d.Periods = append(d.Periods, AIDraftPeriod{
			StartDate:         p.StartDate,
			EndDate:           p.EndDate,
			Price:             p.Price,
			ChildWithBedPrice: p.ChildWithBedPrice,
			ChildNoBedPrice:   p.ChildNoBedPrice,
			InfantPrice:       p.InfantPrice,
			SingleSupplement:  p.SingleSupplement,
			Deposit:           t.DepositPerPerson,
			Note:              p.Note,
		})
	}

	d.InternalNote = internalNote(t)
	return d, warns
}

// internalNote collects facts with no field of their own (staff only).
func internalNote(t *NormalizedTour) string {
	var lines []string
	if t.JoinLandPrice != nil {
		lines = append(lines, fmt.Sprintf("Join Land (ไม่ใช้ตั๋วเครื่องบิน): %s บาท", formatBaht(*t.JoinLandPrice)))
	}
	if t.MinGroupSize != nil {
		lines = append(lines, fmt.Sprintf("กรุ๊ปออกเดินทางขั้นต่ำ %d ท่าน", *t.MinGroupSize))
	}
	lines = append(lines, t.OtherCharges...)
	if len(lines) == 0 {
		return ""
	}
	return "จาก PDF (AI):\n" + strings.Join(lines, "\n")
}

func findCode(items []LookupCode, code, name string) (LookupCode, bool) {
	if c := strings.ToUpper(strings.TrimSpace(code)); c != "" {
		for _, it := range items {
			if strings.EqualFold(it.Code, c) {
				return it, true
			}
		}
	}
	key := matchKey(name)
	if key == "" {
		return LookupCode{}, false
	}
	for _, it := range items {
		if matchKey(it.Name) == key || matchKey(it.Code) == key {
			return it, true
		}
	}
	// "Emirates" vs "Emirates Airlines": one name inside the other.
	if len([]rune(key)) >= 5 {
		for _, it := range items {
			if k := matchKey(it.Name); len([]rune(k)) >= 5 && (strings.Contains(k, key) || strings.Contains(key, k)) {
				return it, true
			}
		}
	}
	return LookupCode{}, false
}

// matchKey lowercases and keeps only letters, digits and Thai marks.
func matchKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func placeName(th, en string) string {
	switch {
	case th != "" && en != "":
		return th + " / " + en
	case th != "":
		return th
	}
	return en
}
