package usecase

import (
	"encoding/json"
	"testing"
	"time"

	"backend/pkg/modules/linecontact/models"
)

func TestCollectContactEvents(t *testing.T) {
	body := `{"destination":"Ubot","events":[
		{"type":"message","timestamp":1762750000000,"source":{"type":"user","userId":"U1"}},
		{"type":"follow","timestamp":1762750001000,"source":{"type":"user","userId":"U2"}},
		{"type":"unfollow","timestamp":1762750002000,"source":{"type":"user","userId":"U3"}},
		{"type":"message","timestamp":1762750003000,"source":{"type":"group","groupId":"G1","userId":"U4"}},
		{"type":"postback","timestamp":1762750004000,"source":{"type":"user","userId":"U5"}},
		{"type":"message","timestamp":1762750005000,"source":{"type":"user"}}
	]}`
	var payload models.WebhookPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatal(err)
	}

	got := collectContactEvents(payload.Events)
	want := []contactEvent{
		{UserID: "U1", Kind: contactMessage, At: time.UnixMilli(1762750000000)},
		{UserID: "U2", Kind: contactFollow, At: time.UnixMilli(1762750001000)},
		{UserID: "U3", Kind: contactUnfollow, At: time.UnixMilli(1762750002000)},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].UserID != want[i].UserID || got[i].Kind != want[i].Kind || !got[i].At.Equal(want[i].At) {
			t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCollectContactEventsEmpty(t *testing.T) {
	// The console's "Verify" button sends no events.
	if got := collectContactEvents(nil); len(got) != 0 {
		t.Fatalf("got %+v, want none", got)
	}
}
