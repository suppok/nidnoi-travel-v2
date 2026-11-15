package usecase

import (
	"context"
	"math"
	"time"

	"backend/pkg/constants/apicode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/dashboard/models"
	"backend/pkg/modules/shared/services/linebookingscope"
)

const (
	defaultDays       = 30
	topLimit          = 8
	seatsWindowDays   = 30 // tours: departures in the next 30 days
	lowSeatsThreshold = 5
	departuresDays    = 14 // LINE bookings: departures in the next 14 days
	departuresLimit   = 10
)

var allowedDays = []any{7, 30, 90, 365}

func (u *dashboardUseCase) SummaryValidate(request *models.DashboardSummaryRequest) error {
	if request.Days == 0 {
		request.Days = defaultDays
	}
	return validation.ValidateStruct(request,
		validation.Field(&request.Days, validation.In(allowedDays...)),
	)
}

// Summary returns the dashboard sections the caller may see. Each section needs its list
// permission; LINE bookings are limited to the caller's own like POST /line-bookings/list.
func (u *dashboardUseCase) Summary(ctx context.Context, request *models.DashboardSummaryRequest) (*models.DashboardSummaryResponse, error) {
	if err := u.SummaryValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	now := time.Now()
	rng := newSummaryRange(now, request.Days)
	res := &models.DashboardSummaryResponse{
		Range: models.DashboardRange{
			From:         rng.From.Format(time.DateOnly),
			To:           rng.To.Format(time.DateOnly),
			Days:         request.Days,
			Unit:         rng.Unit,
			PreviousFrom: rng.PreviousFrom.Format(time.DateOnly),
			PreviousTo:   rng.From.AddDate(0, 0, -1).Format(time.DateOnly),
		},
		GeneratedAt: now,
	}

	roles := request.UserRequestInfo.Roles
	allowed := func(code apicode.Code) bool { return u.coreRegistry.Permissions.Allowed(roles, code) }

	var err error
	if allowed(apicode.TourProgramList) {
		if res.Tours, err = u.tours(ctx, rng.To); err != nil {
			return nil, cerrors.Wrap(err, "dashboard tours")
		}
	}
	if allowed(apicode.BookingList) {
		if res.Bookings, err = u.bookings(ctx, rng); err != nil {
			return nil, cerrors.Wrap(err, "dashboard bookings")
		}
	}
	if allowed(apicode.LineBookingList) {
		if res.LineBookings, err = u.lineBookings(ctx, rng, linebookingscope.For(request.UserRequestInfo)); err != nil {
			return nil, cerrors.Wrap(err, "dashboard line bookings")
		}
	}
	return res, nil
}

func (u *dashboardUseCase) tours(ctx context.Context, today time.Time) (*models.DashboardTours, error) {
	tourRepo := u.repositoryRegistry.TourProgramRepository
	periodRepo := u.repositoryRegistry.TourProgramPeriodRepository
	// Departures from tomorrow on: groups leaving today are already gone.
	from, to := today.AddDate(0, 0, 1), today.AddDate(0, 0, seatsWindowDays)

	byState, err := tourRepo.CountByState(ctx)
	if err != nil {
		return nil, err
	}
	countries, err := tourRepo.CountPublicPerCountry(ctx, topLimit)
	if err != nil {
		return nil, err
	}
	seats, err := periodRepo.SeatStats(ctx, from, to, lowSeatsThreshold)
	if err != nil {
		return nil, err
	}
	periods, err := periodRepo.ListUpcoming(ctx, from, to, topLimit)
	if err != nil {
		return nil, err
	}

	out := &models.DashboardTours{
		ByState:   stateCounts(byState),
		ByCountry: make([]models.DashboardCountryTour, len(countries)),
		Seats: models.DashboardSeats{
			Days:              seatsWindowDays,
			LowSeatsThreshold: lowSeatsThreshold,
			Periods:           seats.Periods,
			SeatsTotal:        seats.SeatsTotal,
			SeatsAvailable:    seats.SeatsAvailable,
			SoldOut:           seats.SoldOut,
			LowSeats:          seats.LowSeats,
		},
		LowSeatPeriods: make([]models.DashboardPeriod, len(periods)),
	}
	for i, c := range countries {
		out.ByCountry[i] = models.DashboardCountryTour{CountryID: c.CountryID, NameTh: c.NameTh, Count: c.Count}
	}
	for i, p := range periods {
		out.LowSeatPeriods[i] = models.DashboardPeriod{
			PeriodID:       p.PeriodID,
			TourProgramID:  p.TourProgramID,
			TourCode:       p.TourCode,
			TourName:       p.TourName,
			StartDate:      p.StartDate.Format(time.DateOnly),
			EndDate:        p.EndDate.Format(time.DateOnly),
			Price:          p.Price,
			SeatsTotal:     p.SeatsTotal,
			SeatsAvailable: p.SeatsAvailable,
		}
	}
	return out, nil
}

func (u *dashboardUseCase) bookings(ctx context.Context, rng summaryRange) (*models.DashboardBookings, error) {
	repo := u.repositoryRegistry.BookingRepository
	byState, err := repo.StatsByState(ctx, rng.Current())
	if err != nil {
		return nil, err
	}
	previous, err := repo.StatsByState(ctx, rng.Previous())
	if err != nil {
		return nil, err
	}
	buckets, err := repo.StatsBuckets(ctx, rng.Current(), rng.Unit)
	if err != nil {
		return nil, err
	}
	newCount, err := repo.CountNew(ctx)
	if err != nil {
		return nil, err
	}
	return &models.DashboardBookings{
		Total:         sumCounts(byState),
		PreviousTotal: sumCounts(previous),
		NewCount:      newCount,
		ByState:       stateCounts(byState),
		Buckets:       fillBuckets(rng, buckets),
	}, nil
}

func (u *dashboardUseCase) lineBookings(ctx context.Context, rng summaryRange, scope *repomodels.LineBookingScope) (*models.DashboardLineBookings, error) {
	repo := u.repositoryRegistry.LineBookingRepository
	cur := repomodels.LineBookingStatsCriteria{Range: rng.Current(), Scope: scope}
	prev := repomodels.LineBookingStatsCriteria{Range: rng.Previous(), Scope: scope}

	totals, err := repo.StatsTotals(ctx, cur)
	if err != nil {
		return nil, err
	}
	previous, err := repo.StatsTotals(ctx, prev)
	if err != nil {
		return nil, err
	}
	buckets, err := repo.StatsBuckets(ctx, cur, rng.Unit)
	if err != nil {
		return nil, err
	}
	byState, err := repo.StatsByState(ctx, cur)
	if err != nil {
		return nil, err
	}
	bySource, err := repo.StatsBySource(ctx, cur, topLimit)
	if err != nil {
		return nil, err
	}
	byTour, err := repo.StatsByTour(ctx, cur, topLimit)
	if err != nil {
		return nil, err
	}
	bySales := []repomodels.LineBookingGroup{}
	if scope == nil {
		if bySales, err = repo.StatsBySales(ctx, cur, topLimit); err != nil {
			return nil, err
		}
	}
	pipeline, err := repo.Pipeline(ctx, scope)
	if err != nil {
		return nil, err
	}
	departures, err := repo.ListDepartures(ctx, scope, rng.To, rng.To.AddDate(0, 0, departuresDays), departuresLimit)
	if err != nil {
		return nil, err
	}

	out := &models.DashboardLineBookings{
		ScopeAll: scope == nil,
		Totals:   lineTotals(totals),
		Previous: lineTotals(previous),
		Buckets:  fillBuckets(rng, buckets),
		ByState:  stateCounts(byState),
		BySource: lineGroups(bySource),
		ByTour:   lineGroups(byTour),
		BySales:  lineGroups(bySales),
		Pipeline: models.DashboardLinePipeline{
			WaitingConfirm:   pipeline.WaitingConfirm,
			WaitingDocuments: pipeline.WaitingDocuments,
			WaitingReview:    pipeline.WaitingReview,
			BalancePending:   pipeline.BalancePending,
		},
		DeparturesDays: departuresDays,
		Departures:     make([]models.DashboardDeparture, len(departures)),
	}
	for i := range departures {
		b := &departures[i]
		d := models.DashboardDeparture{
			ID:           b.ID,
			BookingNo:    b.BookingNo,
			TourCode:     b.TourCode,
			TourName:     b.TourName,
			StartDate:    b.StartDate.Format(time.DateOnly),
			EndDate:      b.EndDate.Format(time.DateOnly),
			CustomerName: b.CustomerName,
			Travellers:   b.Travellers(),
			State:        b.State,
			TotalPrice:   b.TotalPrice,
		}
		if b.SalesUser != nil {
			d.SalesUser = &models.DashboardUser{ID: b.SalesUser.ID, Name: b.SalesUser.Name}
		}
		out.Departures[i] = d
	}
	return out, nil
}

// --- mapping helpers ---

func stateCounts(rows []repomodels.StateCount) []models.DashboardStateCount {
	out := make([]models.DashboardStateCount, len(rows))
	for i, r := range rows {
		out[i] = models.DashboardStateCount{State: r.State, Count: r.Count}
	}
	return out
}

func sumCounts(rows []repomodels.StateCount) int {
	n := 0
	for _, r := range rows {
		n += r.Count
	}
	return n
}

// fillBuckets returns every bucket of the range, zeros where the query had no rows.
func fillBuckets(rng summaryRange, rows []repomodels.BucketCount) []models.DashboardBucket {
	byDate := make(map[string]repomodels.BucketCount, len(rows))
	for _, r := range rows {
		byDate[r.Bucket.Format(time.DateOnly)] = r
	}
	starts := rng.Buckets()
	out := make([]models.DashboardBucket, len(starts))
	for i, d := range starts {
		key := d.Format(time.DateOnly)
		r := byDate[key]
		out[i] = models.DashboardBucket{Date: key, Count: r.Count, Revenue: r.Revenue}
	}
	return out
}

func lineTotals(t *repomodels.LineBookingTotals) models.DashboardLineTotals {
	out := models.DashboardLineTotals{
		Count:      t.Count,
		PaidCount:  t.PaidCount,
		Cancelled:  t.Cancelled,
		Revenue:    t.Revenue,
		Travellers: t.Travellers,
	}
	if t.PaidCount > 0 {
		out.AverageValue = int(math.Round(float64(t.Revenue) / float64(t.PaidCount)))
	}
	return out
}

func lineGroups(rows []repomodels.LineBookingGroup) []models.DashboardLineGroup {
	out := make([]models.DashboardLineGroup, len(rows))
	for i, r := range rows {
		out[i] = models.DashboardLineGroup{
			Key:        r.Key,
			Label:      r.Label,
			Count:      r.Count,
			PaidCount:  r.PaidCount,
			Revenue:    r.Revenue,
			Travellers: r.Travellers,
		}
	}
	return out
}
