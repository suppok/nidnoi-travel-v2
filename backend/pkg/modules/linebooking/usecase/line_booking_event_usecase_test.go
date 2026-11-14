package usecase

import (
	"testing"

	"backend/pkg/domain"
)

func TestDiffSnapshots(t *testing.T) {
	before := &domain.LineBooking{AdultTwinCount: 2, AdultTwinPrice: 10000, TotalPrice: 20000, CustomerName: "A",
		LineContact: &domain.LineContact{DisplayName: "Pocky"}}
	after := *before
	after.AdultTwinCount = 3
	after.TotalPrice = 30000
	after.Charges = []domain.LineBookingCharge{{Kind: "visa", Count: 3, Price: 1500}, {Kind: "option", Name: "ประกัน", Count: 1, Price: 500}}
	after.LineContact = nil

	got := diffSnapshots(dealSnapshot(before), dealSnapshot(&after))
	want := []domain.LineBookingEventChange{
		{Field: "adultTwin", From: "2", To: "3"},
		{Field: "charges", From: "", To: "วีซ่า 3×1500, ประกัน 1×500"},
		{Field: "totalPrice", From: "20000", To: "30000"},
		{Field: "lineContact", From: "Pocky", To: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d changes %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("change %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	if same := diffSnapshots(dealSnapshot(before), dealSnapshot(before)); len(same) != 0 {
		t.Errorf("no edit should have no changes, got %+v", same)
	}
}
