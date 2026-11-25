package domainmodels

import (
	"strconv"
	"strings"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// PublicPeriodSummary is what the customer site bakes into static HTML from a tour's periods
// (cards, JSON-LD). It is comparable: when it changes, the site needs a rebuild; when it doesn't
// (seat changes, prices above the starting price), the live period table is enough.
type PublicPeriodSummary struct {
	StartingPrice               null.Int
	StartingPriceBeforeDiscount null.Int
	NextDepartureDate           string // YYYY-MM-DD, "" without upcoming periods
	PeriodCount                 int
	// FireSales encodes the โปรไฟไหม้ periods (id, order, prices, dates) so any promo change rebuilds the home page.
	FireSales string
}

// SummarizePublicPeriods takes a tour's active upcoming periods in departure order. The starting
// price is the lowest price (sold-out periods included); on a tie the earliest departure wins.
func SummarizePublicPeriods(periods []domain.TourProgramPeriod) PublicPeriodSummary {
	s := PublicPeriodSummary{PeriodCount: len(periods)}
	var cheapest *domain.TourProgramPeriod
	for i := range periods {
		if cheapest == nil || periods[i].Price < cheapest.Price {
			cheapest = &periods[i]
		}
	}
	if cheapest != nil {
		s.StartingPrice = null.IntFrom(int64(cheapest.Price))
		s.StartingPriceBeforeDiscount = cheapest.PriceBeforeDiscount
		s.NextDepartureDate = periods[0].StartDate.Format("2006-01-02")
	}
	var b strings.Builder
	for i := range periods {
		p := &periods[i]
		if !IsFireSale(p) {
			continue
		}
		b.WriteString(p.ID + ":" + strconv.Itoa(p.FireSaleOrder) + ":" + strconv.Itoa(p.Price) + ":" +
			strconv.FormatInt(p.PriceBeforeDiscount.Int64, 10) + ":" + p.StartDate.Format("2006-01-02") + ":" +
			p.EndDate.Format("2006-01-02") + ";")
	}
	s.FireSales = b.String()
	return s
}

// IsFireSale reports whether a period is a โปรไฟไหม้ the customer site may show: flagged and still
// discounted (price_before_discount above price).
func IsFireSale(p *domain.TourProgramPeriod) bool {
	return p.FireSale && p.PriceBeforeDiscount.Valid && p.PriceBeforeDiscount.Int64 > int64(p.Price)
}
