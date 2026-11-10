package usecase

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/tourprogram/models"
)

// Public (customer site) actions. Pure reads: registry repositories, no transaction.
// A period is public when it is active and departs after today (Thailand time).

// publicCodePattern bounds the :code path segment before it reaches SQL; anything else is NOT_FOUND.
var publicCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)

func (u *tourProgramUseCase) PublicListValidate(request *models.TourProgramPublicListRequest) error {
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	request.Keyword = strings.TrimSpace(request.Keyword)
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Country, validation.Slug, validation.RuneLength(0, 60)),
		validation.Field(&request.City, validation.Slug, validation.RuneLength(0, 60)),
		validation.Field(&request.DateFrom, validation.Date(time.DateOnly)),
		validation.Field(&request.DateTo, validation.Date(time.DateOnly)),
		validation.Field(&request.Page, validation.Min(1)),
		validation.Field(&request.PageSize, validation.Min(1), validation.Max(maxPageSize)),
	)
}

func parseOptionalDate(s string) null.Time {
	if s == "" {
		return null.Time{}
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return null.Time{}
	}
	return null.TimeFrom(t)
}

func (u *tourProgramUseCase) PublicList(ctx context.Context, request *models.TourProgramPublicListRequest) (*models.TourProgramPublicListResponse, error) {
	if err := u.PublicListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	today := utils.TodayInThailand(time.Now())
	tours, total, err := u.repositoryRegistry.TourProgramRepository.ListPublic(ctx, repomodels.TourProgramPublicListCriteria{
		Keyword:     request.Keyword,
		CountrySlug: request.Country,
		CitySlug:    request.City,
		DepartAfter: today,
		DateFrom:    parseOptionalDate(request.DateFrom),
		DateTo:      parseOptionalDate(request.DateTo),
		Page:        request.Page,
		PageSize:    request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list public tour programs")
	}

	ids := make([]string, len(tours))
	for i, t := range tours {
		ids[i] = t.ID
	}
	periods, err := u.repositoryRegistry.TourProgramPeriodRepository.ListPublicByTours(ctx, ids, today)
	if err != nil {
		return nil, cerrors.Wrap(err, "list public periods")
	}
	byTour := make(map[string][]domain.TourProgramPeriod, len(tours))
	for _, p := range periods {
		byTour[p.TourProgramID] = append(byTour[p.TourProgramID], p)
	}

	items := make([]models.TourProgramPublicItem, len(tours))
	for i := range tours {
		items[i] = models.NewTourProgramPublicItem(&tours[i], byTour[tours[i].ID])
	}
	return &models.TourProgramPublicListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
