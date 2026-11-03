package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/city"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500
)

type cityUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewCityUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) city.CityUseCase {
	return &cityUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// Shared field rules for create and update.

func slugRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.Length(1, 60), validation.Slug}
}

func nameRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.RuneLength(1, 100)}
}

// refIDRules: optional, but a given ref ID must be positive (Min skips zero values).
func refIDRules() []validation.Rule {
	return []validation.Rule{validation.By(func(value any) error {
		if id, ok := value.(null.Int); ok && id.Valid && id.Int64 < 1 {
			return errors.New("must be a positive number")
		}
		return nil
	})}
}

func descriptionRules() []validation.Rule {
	return []validation.Rule{validation.RuneLength(0, 5000)}
}

func heroImageRules() []validation.Rule {
	return []validation.Rule{validation.Length(0, 500)}
}

func sortOrderRules() []validation.Rule {
	return []validation.Rule{validation.Min(0), validation.Max(100000)}
}

// normalizeText trims display fields; an empty hero image becomes NULL.
func normalizeText(nameTh, nameEn, description *string, heroImage *null.String) {
	*nameTh = strings.TrimSpace(*nameTh)
	*nameEn = strings.TrimSpace(*nameEn)
	*description = strings.TrimSpace(*description)
	if s := strings.TrimSpace(heroImage.String); heroImage.Valid && s != "" {
		*heroImage = null.StringFrom(s)
	} else {
		*heroImage = null.String{}
	}
}

// uniqueConflict maps a unique-index violation to SLUG_TAKEN / REF_ID_TAKEN; other errors pass through.
func uniqueConflict(err error) error {
	switch query.UniqueViolationConstraint(err) {
	case "city_slug_key":
		return cerrors.Conflict(errorcode.SlugTaken)
	case "city_ref_id_key":
		return cerrors.Conflict(errorcode.RefIDTaken)
	}
	return err
}

// activeCountry loads the parent country; a missing or deleted one is a countryId field error.
func (u *cityUseCase) activeCountry(ctx context.Context, appDB query.AppIDB, id string) (*domain.Country, error) {
	found, err := u.repositoryRegistry.CountryRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil || !found.Status {
		return nil, cerrors.FieldError("countryId", "country not found")
	}
	return found, nil
}

// findForUpdate loads the city (with its country) inside a transaction or fails with NOT_FOUND.
func (u *cityUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.City, error) {
	found, err := u.repositoryRegistry.CityRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}

// enqueueIfShown rebuilds the customer site when the city has public tours. paths are the city's
// landing paths (before and after a country move). Called after commit; a failed count is logged.
func (u *cityUseCase) enqueueIfShown(ctx context.Context, cityID string, paths []string) {
	n, err := u.repositoryRegistry.TourProgramRepository.CountPublicByCity(ctx, cityID)
	if err != nil {
		u.coreRegistry.Logger.Error("count public tours for rebuild", zap.Error(err), zap.String("cityId", cityID))
		return
	}
	if n > 0 {
		u.shareRegistry.ServiceRegistry.Revalidator.Enqueue(ctx, paths)
	}
}
