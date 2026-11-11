package usecase

import (
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/linebookingpayment"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

func period() *domain.TourProgramPeriod {
	return &domain.TourProgramPeriod{Price: 30000, SingleSupplement: null.IntFrom(5000), ChildWithBedPrice: null.IntFrom(29000)}
}

func TestQuoteLinePeriodPrices(t *testing.T) {
	// An older LIFF build sends counts only: the period quote, as before.
	b, err := quoteLine(period(), &models.LineBookingCreateRequest{AdultTwin: 2, AdultSingle: 1, ChildWithBed: 1, PaymentMethod: linebookingpayment.Transfer})
	if err != nil {
		t.Fatal(err)
	}
	if want := 2*30000 + 35000 + 29000; b.TotalPrice != want || b.CardFee != 0 {
		t.Fatalf("total = %d fee = %d, want %d and 0", b.TotalPrice, b.CardFee, want)
	}
	if b.AdultSinglePrice.Int64 != 35000 {
		t.Fatalf("single price = %d", b.AdultSinglePrice.Int64)
	}
}

func TestQuoteLineSummaryDeal(t *testing.T) {
	r := &models.LineBookingCreateRequest{
		AdultTwin: 2, AdultSingle: 1, ChildWithBed: 1, Infant: 1,
		AdultPrice: null.IntFrom(29000), ChildWithBedPrice: null.IntFrom(28000), InfantPrice: null.IntFrom(5000),
		DiscountPerPerson: 500,
		Charges: []models.LineBookingChargeReq{
			{Kind: linebookingpayment.ChargeVisa, Count: 2, Price: 1500},
			{Kind: linebookingpayment.ChargeFuelTax, Count: 5, Price: 1000},
			{Kind: linebookingpayment.ChargeOption, Name: "unused", Count: 0, Price: 900},
		},
		PaymentMethod: linebookingpayment.Credit, CreditFeePercent: 3,
		IsDeposit: true, DepositPerPerson: 5000,
	}
	b, err := quoteLine(period(), r)
	if err != nil {
		t.Fatal(err)
	}
	// 2x29000 + (29000+5000) + 28000 + 5000 - 5x500 + 3000 + 5000 = 130500; fee 3% of the 25000 deposit.
	if b.CardFee != 750 || b.TotalPrice != 130500+750 {
		t.Fatalf("fee = %d total = %d", b.CardFee, b.TotalPrice)
	}
	if len(b.Charges) != 2 {
		t.Fatalf("charges = %v, want the unused line dropped", b.Charges)
	}
	if b.CreditFeeBps != 300 {
		t.Fatalf("bps = %d", b.CreditFeeBps)
	}
}

func TestQuoteLineErrors(t *testing.T) {
	if _, err := quoteLine(period(), &models.LineBookingCreateRequest{AdultTwin: 1, ChildNoBed: 1}); err == nil {
		t.Fatal("want an error for a child-no-bed count without a price")
	}
	if _, err := quoteLine(period(), &models.LineBookingCreateRequest{AdultTwin: 1, ChildNoBed: 1, ChildNoBedPrice: null.IntFrom(0)}); err != nil {
		t.Fatalf("an entered price of 0 is allowed: %v", err)
	}
	if _, err := quoteLine(period(), &models.LineBookingCreateRequest{AdultTwin: 1, DiscountPerPerson: 40000}); err == nil {
		t.Fatal("want an error for a discount above the price")
	}
	if _, err := quoteLine(period(), &models.LineBookingCreateRequest{AdultTwin: 1, IsDeposit: true, DepositPerPerson: 40000}); err == nil {
		t.Fatal("want an error for a deposit above the total")
	}
}

func TestCardFee(t *testing.T) {
	for _, tt := range []struct{ amount, bps, want int }{
		{15000, 260, 390}, {96100, 300, 2883}, {1001, 300, 31}, {0, 300, 0},
	} {
		if got := cardFee(tt.amount, tt.bps); got != tt.want {
			t.Errorf("cardFee(%d, %d) = %d, want %d", tt.amount, tt.bps, got, tt.want)
		}
	}
}
