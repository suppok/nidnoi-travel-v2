package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/airline"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500

	// logoPrefix must match the airline-logo key prefix in pkg/modules/media/usecase.
	logoPrefix = "airlines/logos/"
)

// IATA designators are 2 characters (letters or digits, e.g. TG, 3K); ICAO codes are 3 letters.
var codeRe = regexp.MustCompile(`^[A-Z0-9]{2,3}$`)

type airlineUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewAirlineUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) airline.AirlineUseCase {
	return &airlineUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// Shared field rules for create and update.

func refIDRules() []validation.Rule {
	return []validation.Rule{validation.By(func(value any) error {
		if id, ok := value.(null.Int); ok && id.Valid && id.Int64 < 1 {
			return errors.New("must be a positive number")
		}
		return nil
	})}
}

func codeRules() []validation.Rule {
	return []validation.Rule{
		validation.Required,
		validation.Match(codeRe).Error("must be 2-3 letters or digits (IATA or ICAO code)"),
	}
}

func nameRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.RuneLength(1, 200)}
}

// logoRules accepts only URLs this backend uploaded: <R2_PUBLIC_URL>/airlines/logos/<file>.
func (u *airlineUseCase) logoRules() []validation.Rule {
	return []validation.Rule{validation.By(func(value any) error {
		v, _ := value.(null.String)
		if !v.Valid {
			return nil
		}
		base := u.shareRegistry.ClientRegistry.ObjectStorage.PublicBaseURL()
		if base == "" {
			return errors.New("file storage is not configured")
		}
		rest, ok := strings.CutPrefix(v.String, base+"/"+logoPrefix)
		if !ok || rest == "" || strings.ContainsAny(rest, "?#/") || strings.Contains(rest, "..") {
			return errors.New("must be a file uploaded through /media/upload")
		}
		return nil
	})}
}

// normalize uppercases the code, trims text and turns an empty logo into NULL.
func normalize(code, name *string, logo *null.String) {
	*code = strings.ToUpper(strings.TrimSpace(*code))
	*name = strings.TrimSpace(*name)
	if s := strings.TrimSpace(logo.String); logo.Valid && s != "" {
		*logo = null.StringFrom(s)
	} else {
		*logo = null.String{}
	}
}

// uniqueConflict maps a unique-index violation to CODE_TAKEN / REF_ID_TAKEN; other errors pass through.
func uniqueConflict(err error) error {
	switch query.UniqueViolationConstraint(err) {
	case "airline_code_key":
		return cerrors.Conflict(errorcode.CodeTaken)
	case "airline_ref_id_key":
		return cerrors.Conflict(errorcode.RefIDTaken)
	}
	return err
}

// findForUpdate loads the airline inside a transaction or fails with NOT_FOUND.
func (u *airlineUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.Airline, error) {
	found, err := u.repositoryRegistry.AirlineRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}
