package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/wholesalestate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/wholesale"
	"backend/pkg/modules/wholesale/models"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500
	maxContacts     = 20
)

var codeRe = regexp.MustCompile(`^[A-Z0-9]+(-[A-Z0-9]+)*$`)

type wholesaleUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewWholesaleUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) wholesale.WholesaleUseCase {
	return &wholesaleUseCase{
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
		validation.Length(1, 20),
		validation.Match(codeRe).Error("must be letters, digits and single hyphens"),
	}
}

func nameRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.RuneLength(1, 200)}
}

func stateRules() []validation.Rule {
	known := make([]any, len(wholesalestate.All))
	for i, s := range wholesalestate.All {
		known[i] = s
	}
	return []validation.Rule{validation.Required, validation.In(known...)}
}

func contactsRules() []validation.Rule {
	return []validation.Rule{
		validation.Length(0, maxContacts),
		validation.Each(validation.By(func(value any) error {
			c, _ := value.(models.WholesaleContactInput)
			return validation.ValidateStruct(&c,
				validation.Field(&c.Name, validation.Required, validation.RuneLength(1, 100)),
				validation.Field(&c.Phone, validation.Required, validation.RuneLength(1, 50)),
			)
		})),
	}
}

// normalize uppercases the code and trims every text field.
func normalize(code, name *string, contacts []models.WholesaleContactInput) {
	*code = strings.ToUpper(strings.TrimSpace(*code))
	*name = strings.TrimSpace(*name)
	for i := range contacts {
		contacts[i].Name = strings.TrimSpace(contacts[i].Name)
		contacts[i].Phone = strings.TrimSpace(contacts[i].Phone)
	}
}

// uniqueConflict maps a unique-index violation to CODE_TAKEN / REF_ID_TAKEN; other errors pass through.
func uniqueConflict(err error) error {
	switch query.UniqueViolationConstraint(err) {
	case "wholesale_code_key":
		return cerrors.Conflict(errorcode.CodeTaken)
	case "wholesale_ref_id_key":
		return cerrors.Conflict(errorcode.RefIDTaken)
	}
	return err
}

// replaceContacts swaps the wholesale's contacts for inputs (in order) and returns the new rows.
func (u *wholesaleUseCase) replaceContacts(ctx context.Context, appDB query.AppIDB, wholesaleID, actorID string, inputs []models.WholesaleContactInput) ([]domain.WholesaleContact, error) {
	now := time.Now()
	contacts := make([]domain.WholesaleContact, len(inputs))
	for i, in := range inputs {
		contacts[i] = domain.WholesaleContact{
			ID:          utils.NewID(),
			WholesaleID: wholesaleID,
			Name:        in.Name,
			Phone:       in.Phone,
			SortOrder:   i,
			CreateDate:  now,
			UpdateDate:  now,
			CreateBy:    actorID,
			UpdateBy:    actorID,
		}
	}
	err := u.repositoryRegistry.WholesaleContactRepository.WithTx(appDB).ReplaceForWholesale(ctx, wholesaleID, contacts)
	return contacts, err
}

// findForUpdate loads the wholesale inside a transaction or fails with NOT_FOUND.
func (u *wholesaleUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.Wholesale, error) {
	found, err := u.repositoryRegistry.WholesaleRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}
