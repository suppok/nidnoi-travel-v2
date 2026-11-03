package usecase

import (
	"context"
	"time"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/country/models"
)

func (u *countryUseCase) PublicList(ctx context.Context) (*models.CountryPublicListResponse, error) {
	countries, err := u.repositoryRegistry.CountryRepository.ListPublic(ctx)
	if err != nil {
		return nil, cerrors.Wrap(err, "list public countries")
	}
	stats, err := u.repositoryRegistry.TourProgramRepository.PublicCountryStats(ctx, utils.TodayInThailand(time.Now()))
	if err != nil {
		return nil, cerrors.Wrap(err, "public country stats")
	}
	byID := make(map[string]repomodels.TaxonomyPublicStats, len(stats))
	for _, s := range stats {
		byID[s.ID] = s
	}

	items := make([]models.CountryPublicItem, len(countries))
	for i, c := range countries {
		s := byID[c.ID]
		items[i] = models.CountryPublicItem{
			Slug:          c.Slug,
			NameTh:        c.NameTh,
			NameEn:        c.NameEn,
			Description:   c.Description,
			HeroImage:     c.HeroImage,
			SortOrder:     c.SortOrder,
			TourCount:     s.TourCount,
			StartingPrice: s.StartingPrice,
			UpdateDate:    c.UpdateDate,
		}
	}
	return &models.CountryPublicListResponse{Items: items}, nil
}
