package commons

import (
	"strings"
	"testing"
	"time"
)

func sp(s string) *string { return &s }
func ip(n int) *int       { return &n }

func date(d, m int, y *int) AIDate { return AIDate{Day: d, Month: m, Year: y} }

var today = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

func TestResolvePeriodDates(t *testing.T) {
	cases := []struct {
		name       string
		start, end AIDate
		want       [2]string
	}{
		{"BE year on both", date(27, 10, ip(2569)), date(31, 10, ip(2569)), [2]string{"2026-10-27", "2026-10-31"}},
		{"year only on end, crosses new year", date(28, 12, nil), date(1, 1, ip(2570)), [2]string{"2026-12-28", "2027-01-01"}},
		{"year only on end, crosses month", date(28, 10, nil), date(1, 11, ip(2569)), [2]string{"2026-10-28", "2026-11-01"}},
		{"two-digit BE year", date(29, 10, ip(69)), date(4, 11, ip(69)), [2]string{"2026-10-29", "2026-11-04"}},
		{"year only on start", date(26, 11, ip(69)), date(2, 12, nil), [2]string{"2026-11-26", "2026-12-02"}},
		{"Gregorian year", date(5, 11, ip(2026)), date(11, 11, ip(2026)), [2]string{"2026-11-05", "2026-11-11"}},
		{"no year: next occurrence", date(3, 3, nil), date(7, 3, nil), [2]string{"2027-03-03", "2027-03-07"}},
	}
	for _, tc := range cases {
		s, e, ok := resolvePeriodDates(tc.start, tc.end, today)
		got := [2]string{s.Format(time.DateOnly), e.Format(time.DateOnly)}
		if !ok || got != tc.want {
			t.Errorf("%s: got %v ok=%v, want %v", tc.name, got, ok, tc.want)
		}
	}
	if _, _, ok := resolvePeriodDates(date(31, 4, ip(2569)), date(4, 5, ip(2569)), today); ok {
		t.Error("31 April must be rejected")
	}
	if _, _, ok := resolvePeriodDates(date(1, 13, ip(2569)), date(4, 1, ip(2570)), today); ok {
		t.Error("month 13 must be rejected")
	}
}

func TestParsePrice(t *testing.T) {
	for in, want := range map[string]int{"23,999 บาท": 23999, "57,900.-": 57900, "6000": 6000, "1,500 บาท/ทริป/ท่าน": 1500, "8,900.00": 8900} {
		if got := parsePrice(sp(in)); got == nil || *got != want {
			t.Errorf("parsePrice(%q) = %v, want %d", in, got, want)
		}
	}
	for _, in := range []string{"", "ฟรี", "0 บาท"} {
		if got := parsePrice(sp(in)); got != nil {
			t.Errorf("parsePrice(%q) = %d, want nil", in, *got)
		}
	}
}

func TestNormalizePeriodsAndPrices(t *testing.T) {
	x := &AIExtraction{
		Name: sp("  โรแมนติกสวิตแอลป์   อิตาลี สวิตเซอร์แลนด์ "), Days: ip(7), Nights: ip(4),
		Countries: []AIPlace{{NameTh: sp("อิตาลี"), NameEn: sp("Italy")}, {NameTh: sp("อิตาลี"), NameEn: sp("Italy")}},
		Periods: []AIPeriod{
			{Start: date(29, 10, ip(69)), End: date(4, 11, ip(69)), AdultPrice: sp("57,900.-"), ChildNoBed: sp("49,900.-"), Infant: sp("6,000 บาท"), SingleRoom: sp("8,900.-"), SingleRoomKind: "supplement", SourceText: "29 ต.ค. - 4 พ.ย. 69"},
			// the same table printed again on another page
			{Start: date(29, 10, ip(69)), End: date(4, 11, ip(69)), AdultPrice: sp("57,900.-"), SourceText: "29 ต.ค. - 4 พ.ย. 69"},
			{Start: date(5, 11, ip(69)), End: date(11, 11, ip(69)), AdultPrice: sp("57,900.-"), ChildWithBed: sp("57,900"), ChildNoBed: sp("57,900"), Infant: sp("ไม่ระบุ"), SingleRoom: sp("66,800"), SingleRoomKind: "unknown", Label: sp("(Promotion)")},
			{Start: date(5, 12, ip(69)), End: date(1, 12, ip(69)), AdultPrice: sp("1")},
		},
		Extras:     AIExtras{DepositPerPerson: sp("10,000 บาท"), JoinLandPrice: sp("-")},
		Confidence: AIConfidence{Main: "high", Taxonomy: "bogus"},
	}
	n, warns := NormalizeAIExtraction(x, today)

	if n.Name != "โรแมนติกสวิตแอลป์ อิตาลี สวิตเซอร์แลนด์" {
		t.Errorf("name = %q (whitespace not collapsed)", n.Name)
	}
	if len(n.Countries) != 1 {
		t.Errorf("countries = %v, want duplicates dropped", n.Countries)
	}
	if len(n.Periods) != 2 {
		t.Fatalf("periods = %+v, want 2 (duplicate and reversed dates dropped)", n.Periods)
	}
	p0, p1 := n.Periods[0], n.Periods[1]
	if p0.StartDate != "2026-10-29" || *p0.Price != 57900 || *p0.ChildNoBedPrice != 49900 || p0.ChildWithBedPrice != nil || *p0.InfantPrice != 6000 || *p0.SingleSupplement != 8900 {
		t.Errorf("period 0 = %+v", p0)
	}
	if *p1.ChildWithBedPrice != 57900 || *p1.ChildNoBedPrice != 57900 || p1.InfantPrice != nil {
		t.Errorf("child prices / unreadable infant price: %+v", p1)
	}
	if p1.SingleSupplement == nil || *p1.SingleSupplement != 8900 {
		t.Errorf("single room 66,800 printed as a total must become supplement 8,900: %v", p1.SingleSupplement)
	}
	if p1.Note != "(Promotion)" {
		t.Errorf("note = %q", p1.Note)
	}
	if *n.DepositPerPerson != 10000 || n.JoinLandPrice != nil {
		t.Errorf("extras = %v %v", n.DepositPerPerson, n.JoinLandPrice)
	}
	if n.Confidence.Taxonomy != "low" || n.Confidence.Main != "high" {
		t.Errorf("confidence = %+v", n.Confidence)
	}
	all := ""
	for _, w := range warns {
		all += w.Field + ": " + w.Message + "\n"
		if strings.Contains(w.Field, "rice") || strings.Contains(w.Field, "deposit") || strings.Contains(w.Field, "singleSupplement") {
			t.Errorf("prices must never warn: %+v", w)
		}
	}
	for _, want := range []string{"ซ้ำในไฟล์", "วันกลับอยู่ก่อนวันไป"} {
		if !strings.Contains(all, want) {
			t.Errorf("warnings missing %q:\n%s", want, all)
		}
	}
}

func TestNormalizeEmptyAnswer(t *testing.T) {
	n, warns := NormalizeAIExtraction(&AIExtraction{Days: ip(0), Nights: ip(9)}, today)
	if n.Days != nil || n.Nights != nil || len(n.Periods) != 0 || n.Countries == nil || n.Periods == nil {
		t.Fatalf("normalized = %+v", n)
	}
	fields := map[string]bool{}
	for _, w := range warns {
		fields[w.Field] = true
	}
	for _, f := range []string{"name", "days", "nights", "countryIds", "periods"} {
		if !fields[f] {
			t.Errorf("missing warning for %s: %+v", f, warns)
		}
	}
}

func TestResolveAIDraft(t *testing.T) {
	n := &NormalizedTour{
		Name: "เซี่ยงไฮ้ ดิสนีย์แลนด์", Days: ip(5), Nights: ip(3),
		AirlineCode: "VZ", AirlineName: "VIETJET AIR", WholesaleTourCode: "GO3MXP-EK073",
		PrintedTourCode:  "NIDN263430",
		Countries:        []NormalizedPlace{{NameTh: "จีน", NameEn: "China"}, {NameTh: "ดาวอังคาร", NameEn: "Mars"}},
		Cities:           []NormalizedCity{{NameTh: "เซี่ยงไฮ้", NameEn: "Shanghai"}, {NameEn: "Milan"}, {NameEn: "Atlantis"}},
		Periods:          []NormalizedPeriod{{StartDate: "2026-10-27", EndDate: "2026-10-31", Price: ip(23999)}},
		DepositPerPerson: ip(10000), JoinLandPrice: ip(5000), MinGroupSize: ip(10),
		OtherCharges: []string{"กรณีถือหนังสือเดินทางต่างชาติ เก็บเพิ่มท่านละ 2,000 บาท"},
	}
	lookup := AILookup{
		Countries: []LookupCountry{{ID: "c-cn", Slug: "china", NameTh: "จีน", NameEn: "China"}, {ID: "c-it", Slug: "italy", NameTh: "อิตาลี", NameEn: "Italy"}},
		Cities:    []LookupCity{{ID: "ci-sh", CountryID: "c-cn", Slug: "shanghai", NameTh: "เซี่ยงไฮ้", NameEn: "Shanghai"}, {ID: "ci-mi", CountryID: "c-it", Slug: "milan", NameTh: "มิลาน", NameEn: "Milan"}},
		Airlines:  []LookupCode{{ID: "a-vz", Code: "VZ", Name: "Thai Viet Jet Air"}},
		ExistingTourCode: func(code string) bool {
			return code == "NIDN263430"
		},
	}
	d, warns := ResolveAIDraft(n, lookup)
	if strings.Join(d.CountryIDs, ",") != "c-cn" || strings.Join(d.CityIDs, ",") != "ci-sh" {
		t.Errorf("countries %v cities %v", d.CountryIDs, d.CityIDs)
	}
	if d.AirlineID == nil || *d.AirlineID != "a-vz" || d.WholesaleID != nil || d.WholesaleTourCode != "GO3MXP-EK073" {
		t.Errorf("airline %v wholesale %v", d.AirlineID, d.WholesaleID)
	}
	if d.State != DraftStateDraft || d.RefID != nil || d.Description != "" {
		t.Errorf("draft defaults: %+v", d)
	}
	if len(d.Periods) != 1 || *d.Periods[0].Deposit != 10000 || d.Periods[0].SeatsTotal != nil {
		t.Errorf("periods = %+v", d.Periods)
	}
	for _, want := range []string{"Join Land (ไม่ใช้ตั๋วเครื่องบิน): 5,000 บาท", "กรุ๊ปออกเดินทางขั้นต่ำ 10 ท่าน", "หนังสือเดินทางต่างชาติ"} {
		if !strings.Contains(d.InternalNote, want) {
			t.Errorf("internal note missing %q:\n%s", want, d.InternalNote)
		}
	}
	all := ""
	for _, w := range warns {
		all += w.Message + "\n"
	}
	for _, w := range warns {
		if w.Field == "wholesaleId" || w.Field == "periods" {
			t.Errorf("no wholesale or seat warnings: %+v", w)
		}
	}
	for _, want := range []string{"Mars", "Atlantis", "Milan", "NIDN263430"} {
		if !strings.Contains(all, want) {
			t.Errorf("warnings missing %q:\n%s", want, all)
		}
	}
}
