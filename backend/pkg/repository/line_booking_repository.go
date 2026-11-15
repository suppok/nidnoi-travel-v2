package repository

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"backend/pkg/constants/linebookingstate"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type LineBookingRepository interface {
	WithTx(tx query.AppIDB) LineBookingRepository
	Create(ctx context.Context, obj *domain.LineBooking) (*domain.LineBooking, error)
	Update(ctx context.Context, obj *domain.LineBooking, columns ...string) (*domain.LineBooking, error)
	// FindByID loads the booking with its LineContact, SalesUser, FinanceUser, BookingSource and CreateUser.
	FindByID(ctx context.Context, id string) (*domain.LineBooking, error)
	// FindByBookingNo is FindByID by the booking no (exact, upper case).
	FindByBookingNo(ctx context.Context, bookingNo string) (*domain.LineBooking, error)
	// List returns LINE bookings newest first (newest confirmation first with Paid), with their
	// LineContact, SalesUser, FinanceUser and BookingSource.
	List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error)

	// --- Dashboard aggregates: active bookings created in the range, within the scope. ---

	StatsTotals(ctx context.Context, criteria repomodels.LineBookingStatsCriteria) (*repomodels.LineBookingTotals, error)
	// StatsBuckets counts bookings per date_trunc(unit) of the create date in Thailand time (unit day,
	// week or month); Revenue sums the paid ones. Empty buckets are absent.
	StatsBuckets(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, unit string) ([]repomodels.BucketCount, error)
	StatsByState(ctx context.Context, criteria repomodels.LineBookingStatsCriteria) ([]repomodels.StateCount, error)
	// StatsBySource / StatsByTour / StatsBySales group the bookings, highest paid revenue first, then
	// most bookings; at most limit rows.
	StatsBySource(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error)
	StatsByTour(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error)
	StatsBySales(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error)
	// Pipeline counts the active bookings in scope (any create date) waiting on someone.
	Pipeline(ctx context.Context, scope *repomodels.LineBookingScope) (*repomodels.LineBookingPipeline, error)
	// ListDepartures returns the paid active bookings in scope departing in [from, to] (dates), with
	// their SalesUser, by departure date; at most limit rows.
	ListDepartures(ctx context.Context, scope *repomodels.LineBookingScope, from, to time.Time, limit int) ([]domain.LineBooking, error)
}

type lineBookingRepository struct{ query query.DB[domain.LineBooking] }

func NewLineBookingRepository(db query.AppIDB) LineBookingRepository {
	return &lineBookingRepository{query: query.New[domain.LineBooking](db)}
}

func (r *lineBookingRepository) WithTx(tx query.AppIDB) LineBookingRepository {
	return NewLineBookingRepository(tx)
}

func (r *lineBookingRepository) Create(ctx context.Context, obj *domain.LineBooking) (*domain.LineBooking, error) {
	return r.query.Insert(ctx, obj)
}

func (r *lineBookingRepository) Update(ctx context.Context, obj *domain.LineBooking, columns ...string) (*domain.LineBooking, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *lineBookingRepository) FindByID(ctx context.Context, id string) (*domain.LineBooking, error) {
	obj := new(domain.LineBooking)
	err := r.query.NewSelect().Model(obj).Relation("LineContact").Relation("SalesUser").Relation("FinanceUser").Relation("BookingSource").Relation("CreateUser").Where("lb.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *lineBookingRepository) FindByBookingNo(ctx context.Context, bookingNo string) (*domain.LineBooking, error) {
	obj := new(domain.LineBooking)
	err := r.query.NewSelect().Model(obj).Relation("LineContact").Relation("SalesUser").Relation("FinanceUser").Relation("BookingSource").Relation("CreateUser").Where("lb.booking_no = ?", bookingNo).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *lineBookingRepository) List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error) {
	var items []domain.LineBooking
	q := r.query.NewSelect().Model(&items).Relation("LineContact").Relation("SalesUser").Relation("FinanceUser").Relation("BookingSource").Relation("CreateUser")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(lb.booking_no) LIKE ?", pattern).
				WhereOr("lower(lb.customer_name) LIKE ?", pattern).
				WhereOr("lb.customer_phone LIKE ?", pattern).
				WhereOr("lower(lb.tour_code) LIKE ?", pattern).
				WhereOr("lower(lb.line_display_name) = ?", strings.ToLower(kw)) // whole LINE name only
		})
	}
	if criteria.State != 0 {
		q = q.Where("lb.state = ?", criteria.State)
	}
	if criteria.Status != nil {
		q = q.Where("lb.status = ?", *criteria.Status)
	}
	if criteria.Paid {
		q = q.Where("lb.state IN (?)", bun.In(linebookingstate.Paid))
	}
	q = whereLineBookingScope(q, criteria.Scope)
	if criteria.FinanceUserID != "" {
		q = q.Where("lb.finance_user_id = ?", criteria.FinanceUserID)
	}
	if criteria.SalesUserID != "" {
		q = q.Where("lb.sales_user_id = ?", criteria.SalesUserID)
	}
	if criteria.LineContactID != "" {
		q = q.Where("lb.line_contact_id = ?", criteria.LineContactID)
	}
	if d := criteria.DocumentsDone; d != nil {
		if *d {
			q = q.Where("lb.documents_at IS NOT NULL")
		} else {
			q = q.Where("lb.documents_at IS NULL")
		}
	}

	if a := criteria.DocumentsApproved; a != nil {
		if *a {
			q = q.Where("lb.documents_approved_at IS NOT NULL")
		} else {
			q = q.Where("lb.documents_approved_at IS NULL")
		}
	}

	if p := criteria.BalancePending; p != nil {
		pending := "lb.state = ? AND lb.balance_submitted_at IS NOT NULL AND lb.balance_confirmed_at IS NULL"
		if *p {
			q = q.Where(pending, linebookingstate.DepositPaid)
		} else {
			q = q.Where("NOT ("+pending+")", linebookingstate.DepositPaid)
		}
	}

	order := "lb.create_date DESC, lb.id ASC"
	switch {
	// Waiting for finance: newest confirmation or fix request first.
	case criteria.DocumentsDone != nil && !*criteria.DocumentsDone && criteria.Paid:
		order = "GREATEST(lb.confirmed_at, lb.fix_requested_at, lb.balance_confirmed_at) DESC NULLS LAST, lb.id ASC"
	case criteria.DocumentsDone != nil && *criteria.DocumentsDone:
		order = "lb.documents_at DESC, lb.id ASC"
	case criteria.Paid:
		order = "lb.confirmed_at DESC NULLS LAST, lb.id ASC"
	}
	total, err := q.
		OrderExpr(order).
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}

// whereLineBookingScope limits q to the bookings the scope may see (nil = every booking).
func whereLineBookingScope(q *bun.SelectQuery, s *repomodels.LineBookingScope) *bun.SelectQuery {
	if s == nil {
		return q
	}
	return q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Where("FALSE")
		if s.Sales {
			q = q.WhereOr("lb.sales_user_id = ?", s.UserID)
		}
		if s.Finance {
			q = q.WhereOr("lb.finance_user_id = ?", s.UserID)
		}
		return q
	})
}

// Paid-only aggregates for the dashboard.
const (
	lineBookingTravellersExpr = "(lb.adult_twin_count + lb.adult_single_count + lb.child_with_bed_count + lb.child_no_bed_count + lb.infant_count)"
	lineBookingPaidFilter     = "FILTER (WHERE lb.state IN (?))"
)

func (r *lineBookingRepository) statsQuery(criteria repomodels.LineBookingStatsCriteria) *bun.SelectQuery {
	q := r.query.NewSelect().
		Where("lb.status = TRUE").
		Where("lb.create_date >= ?", criteria.Range.From).
		Where("lb.create_date < ?", criteria.Range.To)
	return whereLineBookingScope(q, criteria.Scope)
}

// groupColumns adds the count, paid count, paid revenue and paid travellers columns.
func groupColumns(q *bun.SelectQuery) *bun.SelectQuery {
	paid := bun.In(linebookingstate.Paid)
	return q.
		ColumnExpr("count(*) AS count").
		ColumnExpr("count(*) "+lineBookingPaidFilter+" AS paid_count", paid).
		ColumnExpr("coalesce(sum(lb.total_price) "+lineBookingPaidFilter+", 0) AS revenue", paid).
		ColumnExpr("coalesce(sum("+lineBookingTravellersExpr+") "+lineBookingPaidFilter+", 0) AS travellers", paid)
}

func (r *lineBookingRepository) StatsTotals(ctx context.Context, criteria repomodels.LineBookingStatsCriteria) (*repomodels.LineBookingTotals, error) {
	out := new(repomodels.LineBookingTotals)
	err := groupColumns(r.statsQuery(criteria)).
		ColumnExpr("count(*) FILTER (WHERE lb.state = ?) AS cancelled", linebookingstate.Cancelled).
		Scan(ctx, out)
	return out, err
}

func (r *lineBookingRepository) StatsBuckets(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, unit string) ([]repomodels.BucketCount, error) {
	var rows []repomodels.BucketCount
	err := r.statsQuery(criteria).
		ColumnExpr("date_trunc(?, lb.create_date AT TIME ZONE 'Asia/Bangkok')::date AS bucket", unit).
		ColumnExpr("count(*) AS count").
		ColumnExpr("coalesce(sum(lb.total_price) "+lineBookingPaidFilter+", 0) AS revenue", bun.In(linebookingstate.Paid)).
		GroupExpr("1").
		OrderExpr("1").
		Scan(ctx, &rows)
	return rows, err
}

func (r *lineBookingRepository) StatsByState(ctx context.Context, criteria repomodels.LineBookingStatsCriteria) ([]repomodels.StateCount, error) {
	var rows []repomodels.StateCount
	err := r.statsQuery(criteria).
		ColumnExpr("lb.state AS state").
		ColumnExpr("count(*) AS count").
		GroupExpr("lb.state").
		OrderExpr("lb.state").
		Scan(ctx, &rows)
	return rows, err
}

func (r *lineBookingRepository) statsGroup(ctx context.Context, q *bun.SelectQuery, limit int) ([]repomodels.LineBookingGroup, error) {
	var rows []repomodels.LineBookingGroup
	err := groupColumns(q).
		GroupExpr("1, 2").
		OrderExpr("revenue DESC, count DESC, 2 ASC").
		Limit(int64(limit)).
		Scan(ctx, &rows)
	return rows, err
}

func (r *lineBookingRepository) StatsBySource(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error) {
	q := r.statsQuery(criteria).
		Join("LEFT JOIN booking_source AS bs ON bs.id = lb.booking_source_id").
		ColumnExpr("coalesce(bs.id::text, '') AS key").
		ColumnExpr("coalesce(bs.name, '') AS label")
	return r.statsGroup(ctx, q, limit)
}

func (r *lineBookingRepository) StatsByTour(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error) {
	// The latest snapshot name per code labels the row (a tour may have been renamed).
	q := r.statsQuery(criteria).
		ColumnExpr("lb.tour_code AS key").
		ColumnExpr("(array_agg(lb.tour_name ORDER BY lb.create_date DESC))[1] AS label")
	var rows []repomodels.LineBookingGroup
	err := groupColumns(q).
		GroupExpr("lb.tour_code").
		OrderExpr("revenue DESC, count DESC, lb.tour_code ASC").
		Limit(int64(limit)).
		Scan(ctx, &rows)
	return rows, err
}

func (r *lineBookingRepository) StatsBySales(ctx context.Context, criteria repomodels.LineBookingStatsCriteria, limit int) ([]repomodels.LineBookingGroup, error) {
	q := r.statsQuery(criteria).
		Join("LEFT JOIN app_user AS su ON su.id = lb.sales_user_id").
		ColumnExpr("coalesce(su.id::text, '') AS key").
		ColumnExpr("coalesce(su.name, '') AS label")
	return r.statsGroup(ctx, q, limit)
}

func (r *lineBookingRepository) Pipeline(ctx context.Context, scope *repomodels.LineBookingScope) (*repomodels.LineBookingPipeline, error) {
	out := new(repomodels.LineBookingPipeline)
	paid := bun.In(linebookingstate.Paid)
	q := r.query.NewSelect().
		Where("lb.status = TRUE").
		ColumnExpr("count(*) FILTER (WHERE lb.state = ?) AS waiting_confirm", linebookingstate.New).
		// A fix request and a confirmed balance clear documents_at: finance owes the next round.
		ColumnExpr("count(*) FILTER (WHERE lb.state IN (?) AND lb.documents_at IS NULL) AS waiting_documents", paid).
		ColumnExpr("count(*) FILTER (WHERE lb.state IN (?) AND lb.documents_at IS NOT NULL AND lb.documents_approved_at IS NULL) AS waiting_review", paid).
		ColumnExpr("count(*) FILTER (WHERE lb.state = ? AND lb.balance_submitted_at IS NOT NULL AND lb.balance_confirmed_at IS NULL) AS balance_pending",
			linebookingstate.DepositPaid)
	err := whereLineBookingScope(q, scope).Scan(ctx, out)
	return out, err
}

func (r *lineBookingRepository) ListDepartures(ctx context.Context, scope *repomodels.LineBookingScope, from, to time.Time, limit int) ([]domain.LineBooking, error) {
	var items []domain.LineBooking
	q := r.query.NewSelect().Model(&items).Relation("SalesUser").
		Where("lb.status = TRUE").
		Where("lb.state IN (?)", bun.In(linebookingstate.Paid)).
		Where("lb.start_date >= ?", from.Format(time.DateOnly)).
		Where("lb.start_date <= ?", to.Format(time.DateOnly))
	err := whereLineBookingScope(q, scope).
		OrderExpr("lb.start_date ASC, lb.booking_no ASC").
		Limit(int64(limit)).
		Scan(ctx)
	return items, err
}
