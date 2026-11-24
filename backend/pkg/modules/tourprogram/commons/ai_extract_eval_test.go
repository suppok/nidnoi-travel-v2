//go:build aieval

// AI extraction eval: reads every PDF in backend/ex with the real Claude API and compares the
// normalized result with its golden file (<name>.expected.json). It costs money (roughly 7-10 baht
// per PDF with Sonnet), so it only builds with the aieval tag and never runs in CI:
//
//	cd backend
//	AI_EVAL_API_KEY=sk-ant-... go test -tags aieval -run TestAIExtractEval -timeout 30m -v ./pkg/modules/tourprogram/commons/
//
// Optional: AI_EVAL_MODEL (default claude-sonnet-5-5), AI_EVAL_ONLY=nidn265004 (one file),
// AI_EVAL_OUT=path.json (save the raw answers and scores). PDFs without a golden file are skipped.
// Files in AIExtractFewShotFiles are reported apart from held-out files: they are the prompt's own
// examples, so they overstate accuracy.
package commons

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/core/config"
	"backend/pkg/modules/shared/client/anthropic"
)

const evalDir = "../../../../ex"

type evalCheck struct {
	Field  string `json:"field"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

type evalFile struct {
	Name     string          `json:"name"`
	HeldOut  bool            `json:"heldOut"`
	Checks   []evalCheck     `json:"checks"`
	Usage    anthropic.Usage `json:"usage"`
	Seconds  float64         `json:"seconds"`
	Answer   json.RawMessage `json:"answer,omitempty"`
	Warnings []AIWarning     `json:"warnings"`
	Error    string          `json:"error,omitempty"`
}

func TestAIExtractEval(t *testing.T) {
	key := os.Getenv("AI_EVAL_API_KEY")
	if key == "" {
		t.Skip("set AI_EVAL_API_KEY to run the AI eval (it calls the Claude API and costs money)")
	}
	model := os.Getenv("AI_EVAL_MODEL")
	if model == "" {
		model = "claude-sonnet-5-5"
	}
	pdfs, _ := filepath.Glob(filepath.Join(evalDir, "*.pdf"))
	sort.Strings(pdfs)
	client := anthropic.New(config.AIConfig{RequestTimeout: 5 * time.Minute, MaxConcurrent: 1, MaxRetries: 3}, zap.NewNop())
	// The golden files hold dates as printed; "today" only matters for rows without a year.
	today := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	var results []evalFile
	for _, pdfPath := range pdfs {
		name := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
		if only := os.Getenv("AI_EVAL_ONLY"); only != "" && only != name {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(evalDir, name+".expected.json"))
		if err != nil {
			t.Logf("%s: no golden file, skipped", name)
			continue
		}
		var want NormalizedTour
		if err := json.Unmarshal(raw, &want); err != nil {
			t.Fatalf("%s.expected.json: %v", name, err)
		}
		pdf, err := os.ReadFile(pdfPath)
		if err != nil {
			t.Fatal(err)
		}

		res := evalFile{Name: name, HeldOut: !slices.Contains(AIExtractFewShotFiles, name)}
		started := time.Now()
		out, err := client.ExtractJSON(context.Background(), anthropic.ExtractRequest{
			APIKey: key, Model: model, PDF: pdf, Instruction: AIExtractInstruction, Schema: AIExtractSchema,
			System: []anthropic.SystemBlock{{Text: AIExtractSystemPrompt, Cache: true}}, MaxTokens: 32000,
		})
		res.Seconds = time.Since(started).Seconds()
		if err != nil {
			res.Error = err.Error()
			t.Errorf("%s: %v", name, err)
			results = append(results, res)
			continue
		}
		res.Usage, res.Answer = out.Usage, out.JSON
		var x AIExtraction
		if err := json.Unmarshal(out.JSON, &x); err != nil {
			res.Error = "answer does not decode: " + err.Error()
			t.Errorf("%s: %s", name, res.Error)
			results = append(results, res)
			continue
		}
		got, warns := NormalizeAIExtraction(&x, today)
		res.Warnings = warns
		res.Checks = compareTours(&want, got)
		results = append(results, res)
	}
	if len(results) == 0 {
		t.Skip("no PDFs with golden files in backend/ex")
	}
	t.Log("\n" + evalReport(results, model))
	if path := os.Getenv("AI_EVAL_OUT"); path != "" {
		b, _ := json.MarshalIndent(map[string]any{"model": model, "promptVersion": AIExtractPromptVersion, "files": results}, "", "  ")
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Errorf("write %s: %v", path, err)
		}
	}
}

// compareTours scores one file. Free-text fields (highlight) are not scored.
func compareTours(want, got *NormalizedTour) []evalCheck {
	var cs []evalCheck
	eq := func(field string, w, g any) {
		cs = append(cs, evalCheck{Field: field, Pass: fmt.Sprint(w) == fmt.Sprint(g), Detail: fmt.Sprintf("want %v, got %v", w, g)})
	}
	eq("name", want.Name, got.Name)
	eq("days", deref(want.Days), deref(got.Days))
	eq("nights", deref(want.Nights), deref(got.Nights))
	eq("airlineCode", want.AirlineCode, got.AirlineCode)
	eq("wholesaleTourCode", strings.ToUpper(want.WholesaleTourCode), strings.ToUpper(got.WholesaleTourCode))
	eq("printedTourCode", want.PrintedTourCode, got.PrintedTourCode)

	cs = append(cs, setCheck("countries", names(want.Countries), names(got.Countries), true))
	cs = append(cs, setCheck("cities", cityNames(want.Cities), cityNames(got.Cities), false))

	// Periods: matched by dates; each matched period scores its prices and note.
	gotBy := map[string]NormalizedPeriod{}
	for _, p := range got.Periods {
		gotBy[p.StartDate+"/"+p.EndDate] = p
	}
	wantKeys := map[string]bool{}
	for _, w := range want.Periods {
		k := w.StartDate + "/" + w.EndDate
		wantKeys[k] = true
		g, ok := gotBy[k]
		cs = append(cs, evalCheck{Field: "periods.found", Pass: ok, Detail: k})
		if !ok {
			continue
		}
		eq("periods.price", deref(w.Price), deref(g.Price))
		eq("periods.childWithBedPrice", deref(w.ChildWithBedPrice), deref(g.ChildWithBedPrice))
		eq("periods.childNoBedPrice", deref(w.ChildNoBedPrice), deref(g.ChildNoBedPrice))
		eq("periods.infantPrice", deref(w.InfantPrice), deref(g.InfantPrice))
		eq("periods.singleSupplement", deref(w.SingleSupplement), deref(g.SingleSupplement))
		eq("periods.note", matchKey(w.Note), matchKey(g.Note))
	}
	for _, g := range got.Periods {
		if k := g.StartDate + "/" + g.EndDate; !wantKeys[k] {
			cs = append(cs, evalCheck{Field: "periods.extra", Pass: false, Detail: "not in the golden file: " + k})
		}
	}

	eq("depositPerPerson", deref(want.DepositPerPerson), deref(got.DepositPerPerson))
	eq("joinLandPrice", deref(want.JoinLandPrice), deref(got.JoinLandPrice))
	eq("minGroupSize", deref(want.MinGroupSize), deref(got.MinGroupSize))
	eq("governmentShopStops", derefBool(want.GovernmentShopStops), derefBool(got.GovernmentShopStops))
	// Itinerary: the number of days, then each golden place found on the same day (itemFound).
	eq("itinerary.days", len(want.Itinerary), len(got.Itinerary))
	for i, w := range want.Itinerary {
		var places []string
		if i < len(got.Itinerary) {
			places = got.Itinerary[i].Places
		}
		for _, place := range w.Places {
			cs = append(cs, evalCheck{Field: "itinerary.place", Pass: itemFound(place, places), Detail: fmt.Sprintf("day %d: %s", w.Day, place)})
		}
	}
	// Included / excluded lists are free text: scored on having found the list, and on each golden
	// item having a matching returned line (itemFound).
	for _, l := range []struct {
		field     string
		want, got []string
	}{{"includedItems", want.IncludedItems, got.IncludedItems}, {"excludedItems", want.ExcludedItems, got.ExcludedItems}} {
		cs = append(cs, evalCheck{Field: l.field + ".found", Pass: (len(l.want) == 0) == (len(l.got) == 0),
			Detail: fmt.Sprintf("want %d items, got %d", len(l.want), len(l.got))})
		for _, w := range l.want {
			cs = append(cs, evalCheck{Field: l.field + ".item", Pass: itemFound(w, l.got), Detail: w})
		}
		for _, g := range l.got {
			if noise := noiseItem(g); noise != "" {
				cs = append(cs, evalCheck{Field: l.field + ".noise", Pass: false, Detail: fmt.Sprintf("%q looks like boilerplate (%s)", g, noise)})
			}
		}
	}
	// otherCharges: every amount in the golden lines must appear in some returned line.
	gotText := strings.Join(got.OtherCharges, " | ")
	for _, line := range want.OtherCharges {
		amount := regexp.MustCompile(`[\d,]+`).FindString(strings.ReplaceAll(line, "(Business class)", ""))
		cs = append(cs, evalCheck{Field: "otherCharges", Pass: amount != "" && strings.Contains(gotText, amount), Detail: line})
	}
	return cs
}

// setCheck: one check per list; passes when every wanted name is found (and, when strict, nothing extra).
func setCheck(field string, want, got []string, strict bool) evalCheck {
	missing, extra := diff(want, got), diff(got, want)
	pass := len(missing) == 0 && (!strict || len(extra) == 0)
	return evalCheck{Field: field, Pass: pass, Detail: fmt.Sprintf("missing %v, extra %v", missing, extra)}
}

func diff(a, b []string) []string {
	var out []string
	for _, x := range a {
		if !slices.Contains(b, x) {
			out = append(out, x)
		}
	}
	return out
}

func names(ps []NormalizedPlace) []string {
	var out []string
	for _, p := range ps {
		out = append(out, matchKey(p.NameEn))
	}
	return out
}

func cityNames(ps []NormalizedCity) []string {
	var out []string
	for _, p := range ps {
		out = append(out, matchKey(strings.NewReplacer("è", "e", "é", "e").Replace(p.NameEn)))
	}
	return out
}

func derefBool(p *bool) any {
	if p == nil {
		return "null"
	}
	return *p
}

func deref(p *int) any {
	if p == nil {
		return "null"
	}
	return *p
}

// evalReport prints accuracy per field (held-out and few-shot files apart) and per file.
func evalReport(results []evalFile, model string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "AI extraction eval — model %s, prompt %s\n", model, AIExtractPromptVersion)
	for _, group := range []struct {
		title   string
		heldOut bool
	}{{"HELD-OUT files (not in the prompt)", true}, {"FEW-SHOT files (the prompt's own examples)", false}} {
		pass, total := map[string]int{}, map[string]int{}
		var fields []string
		files := 0
		for _, r := range results {
			if r.HeldOut != group.heldOut {
				continue
			}
			files++
			for _, c := range r.Checks {
				if _, ok := total[c.Field]; !ok {
					fields = append(fields, c.Field)
				}
				total[c.Field]++
				if c.Pass {
					pass[c.Field]++
				}
			}
		}
		if files == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s: %d file(s)\n", group.title, files)
		allPass, allTotal := 0, 0
		for _, f := range fields {
			allPass += pass[f]
			allTotal += total[f]
			fmt.Fprintf(&b, "  %-28s %3d/%-3d %5.1f%%\n", f, pass[f], total[f], 100*float64(pass[f])/float64(total[f]))
		}
		fmt.Fprintf(&b, "  %-28s %3d/%-3d %5.1f%%\n", "ALL CHECKS", allPass, allTotal, 100*float64(allPass)/float64(max(allTotal, 1)))
	}
	b.WriteString("\nPer file:\n")
	for _, r := range results {
		tag := "few-shot"
		if r.HeldOut {
			tag = "held-out"
		}
		if r.Error != "" {
			fmt.Fprintf(&b, "  %s (%s): ERROR %s\n", r.Name, tag, r.Error)
			continue
		}
		pass := 0
		var failed []string
		for _, c := range r.Checks {
			if c.Pass {
				pass++
			} else {
				failed = append(failed, fmt.Sprintf("%s (%s)", c.Field, c.Detail))
			}
		}
		fmt.Fprintf(&b, "  %s (%s): %d/%d checks, %.0fs, %d in / %d out tokens, %d warnings\n",
			r.Name, tag, pass, len(r.Checks), r.Seconds, r.Usage.InputTokens, r.Usage.OutputTokens, len(r.Warnings))
		for _, f := range failed {
			fmt.Fprintf(&b, "      ✗ %s\n", f)
		}
	}
	return b.String()
}
