package repository

import (
	"context"
	"strings"

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
	// FindByID loads the booking with its LineContact, SalesUser, FinanceUser and BookingSource.
	FindByID(ctx context.Context, id string) (*domain.LineBooking, error)
	// List returns LINE bookings newest first (newest confirmation first with Paid), with their
	// LineContact, SalesUser, FinanceUser and BookingSource.
	List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error)
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
	err := r.query.NewSelect().Model(obj).Relation("LineContact").Relation("SalesUser").Relation("FinanceUser").Relation("BookingSource").Where("lb.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *lineBookingRepository) List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error) {
	var items []domain.LineBooking
	q := r.query.NewSelect().Model(&items).Relation("LineContact").Relation("SalesUser").Relation("FinanceUser").Relation("BookingSource")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(lb.booking_no) LIKE ?", pattern).
				WhereOr("lower(lb.customer_name) LIKE ?", pattern).
				WhereOr("lb.customer_phone LIKE ?", pattern).
				WhereOr("lower(lb.tour_code) LIKE ?", pattern).
				WhereOr("lower(lb.line_display_name) LIKE ?", pattern)
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
	if s := criteria.Scope; s != nil {
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
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
	if criteria.FinanceUserID != "" {
		q = q.Where("lb.finance_user_id = ?", criteria.FinanceUserID)
	}
	if criteria.SalesUserID != "" {
		q = q.Where("lb.sales_user_id = ?", criteria.SalesUserID)
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

	order := "lb.create_date DESC, lb.id ASC"
	switch {
	// Waiting for finance: newest confirmation or fix request first.
	case criteria.DocumentsDone != nil && !*criteria.DocumentsDone && criteria.Paid:
		order = "GREATEST(lb.confirmed_at, lb.fix_requested_at) DESC NULLS LAST, lb.id ASC"
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
