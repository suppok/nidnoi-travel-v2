package usecase

import (
	"context"
	"time"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/city/models"
)

func (u *cityUseCase) PublicList(ctx context.Context) (*models.CityPublicListResponse, error) {
	cities, err := u.repositoryRegistry.CityRepository.ListPublic(ctx)
	if err != nil {
		return nil, cerrors.Wrap(err, "list public cities")
	}
	stats, err := u.repositoryRegistry.TourProgramRepository.PublicCityStats(ctx, utils.TodayInThailand(time.Now()))
	if err != nil {
		return nil, cerrors.Wrap(err, "public city stats")
	}
	byID := make(map[string]repomodels.TaxonomyPublicStats, len(stats))
	for _, s := range stats {
		byID[s.ID] = s
	}

	items := make([]models.CityPublicItem, 0, len(cities))
	for _, c := range cities {
		if c.Country == nil {
			continue
		}
		s := byID[c.ID]
		items = append(items, models.CityPublicItem{
			Slug:          c.Slug,
			CountrySlug:   c.Country.Slug,
			NameTh:        c.NameTh,
			NameEn:        c.NameEn,
			Description:   c.Description,
			HeroImage:     c.HeroImage,
			SortOrder:     c.SortOrder,
			TourCount:     s.TourCount,
			StartingPrice: s.StartingPrice,
			UpdateDate:    c.UpdateDate,
		})
	}
	return &models.CityPublicListResponse{Items: items}, nil
}
