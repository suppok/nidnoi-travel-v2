package commons

import (
	"encoding/json"
	"testing"
)

// Structured outputs reject a schema with more than 16 union-typed parameters (type arrays or
// anyOf) with a 400. The schema keeps none: "not found" is "", 0 or -1.
func TestAIExtractSchemaHasNoUnionTypes(t *testing.T) {
	var schema any
	if err := json.Unmarshal(AIExtractSchema, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	var unions []string
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch n := v.(type) {
		case map[string]any:
			if _, ok := n["type"].([]any); ok {
				unions = append(unions, path)
			}
			if _, ok := n["anyOf"]; ok {
				unions = append(unions, path)
			}
			for k, child := range n {
				walk(path+"."+k, child)
			}
		case []any:
			for _, child := range n {
				walk(path+"[]", child)
			}
		}
	}
	walk("$", schema)
	if len(unions) > 0 {
		t.Fatalf("union-typed parameters (limit 16, keep 0): %v", unions)
	}
}

// The schema's "not found" values must come out of post-processing as missing.
func TestNormalizeNotFoundValues(t *testing.T) {
	empty, zero, minusOne := "", 0, -1
	x := &AIExtraction{
		Name: &empty, Days: &zero, Nights: &minusOne, PrintedTourCode: &empty,
		Airline: AIAirline{IATACode: &empty, Name: &empty},
		Extras:  AIExtras{DepositPerPerson: &empty, MinGroupSize: &zero},
		Periods: []AIPeriod{{Start: AIDate{Day: 28, Month: 12, Year: &zero}, End: AIDate{Day: 1, Month: 1, Year: ip(2570)}, AdultPrice: sp("27,999"), ChildWithBed: &empty, ChildNoBed: &empty, Infant: &empty, SingleRoom: &empty, Label: &empty}},
	}
	n, _ := NormalizeAIExtraction(x, today)
	if n.Name != "" || n.Days != nil || n.Nights != nil || n.AirlineCode != "" || n.DepositPerPerson != nil || n.MinGroupSize != nil {
		t.Fatalf("not-found values leaked: %+v", n)
	}
	if len(n.Periods) != 1 || n.Periods[0].StartDate != "2026-12-28" || n.Periods[0].EndDate != "2027-01-01" {
		t.Fatalf("year 0 on the start date must take the end date's year: %+v", n.Periods)
	}
	if p := n.Periods[0]; p.ChildWithBedPrice != nil || p.ChildNoBedPrice != nil || p.InfantPrice != nil || p.SingleSupplement != nil || p.Note != "" {
		t.Fatalf("empty prices must stay empty: %+v", p)
	}
}
