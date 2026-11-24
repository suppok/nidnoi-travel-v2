// Package aimodel lists the Claude models an admin can pick for AI features (AI_SETTINGS.model).
package aimodel

// Model is a selectable Claude model.
type Model struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Default is the latest Sonnet: the balance of accuracy and cost for PDF extraction.
const Default = "claude-sonnet-5-5"

// All lists the selectable models, default first. IDs are exact API model IDs (no date suffix).
var All = []Model{
	{ID: "claude-sonnet-5-5", Label: "Claude Sonnet 5.5 (แนะนำ)"},
	{ID: "claude-opus-5-5", Label: "Claude Opus 5.5 (แม่นยำกว่า ราคาสูงกว่า)"},
	{ID: "claude-haiku-5-5", Label: "Claude Haiku 5.5 (เร็วและถูกที่สุด)"},
}

// Valid reports whether id is a selectable model.
func Valid(id string) bool {
	for _, m := range All {
		if m.ID == id {
			return true
		}
	}
	return false
}

// IDs returns the selectable model IDs.
func IDs() []any {
	out := make([]any, len(All))
	for i, m := range All {
		out[i] = m.ID
	}
	return out
}
