package usecase

import (
	"context"
	"strings"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/bookingsource"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500
)

type bookingSourceUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewBookingSourceUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) bookingsource.BookingSourceUseCase {
	return &bookingSourceUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// Shared field rules for create and update.

func nameRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.RuneLength(1, 100)}
}

func shortNameRules() []validation.Rule {
	return []validation.Rule{validation.RuneLength(0, 50)}
}

func sortOrderRules() []validation.Rule {
	return []validation.Rule{validation.Min(0), validation.Max(9999)}
}

func normalize(name, shortName *string) {
	*name = strings.TrimSpace(*name)
	*shortName = strings.TrimSpace(*shortName)
}

// uniqueConflict maps the name unique-index violation to NAME_TAKEN; other errors pass through.
func uniqueConflict(err error) error {
	if query.UniqueViolationConstraint(err) == "booking_source_name_key" {
		return cerrors.Conflict(errorcode.NameTaken)
	}
	return err
}

// findForUpdate loads the booking source inside a transaction or fails with NOT_FOUND.
func (u *bookingSourceUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.BookingSource, error) {
	found, err := u.repositoryRegistry.BookingSourceRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}
