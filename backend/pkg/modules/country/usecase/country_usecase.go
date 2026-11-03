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
	"backend/pkg/modules/country"
	"backend/pkg/modules/shared/services/revalidation"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500
)

type countryUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewCountryUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) country.CountryUseCase {
	return &countryUseCase{
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
	case "country_slug_key":
		return cerrors.Conflict(errorcode.SlugTaken)
	case "country_ref_id_key":
		return cerrors.Conflict(errorcode.RefIDTaken)
	}
	return err
}

// findForUpdate loads the country inside a transaction or fails with NOT_FOUND.
func (u *countryUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.Country, error) {
	found, err := u.repositoryRegistry.CountryRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}

// enqueueIfShown rebuilds the customer site when the country has public tours (otherwise it appears
// on no page). Called after the transaction commits; a failed count is logged, never returned.
func (u *countryUseCase) enqueueIfShown(ctx context.Context, c *domain.Country) {
	n, err := u.repositoryRegistry.TourProgramRepository.CountPublicByCountry(ctx, c.ID)
	if err != nil {
		u.coreRegistry.Logger.Error("count public tours for rebuild", zap.Error(err), zap.String("countryId", c.ID))
		return
	}
	if n > 0 {
		u.shareRegistry.ServiceRegistry.Revalidator.Enqueue(ctx, revalidation.PathsForCountry(c))
	}
}
