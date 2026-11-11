package usecase

import (
	"context"
	"slices"

	"backend/pkg/constants/role"
	"backend/pkg/domain"
	"backend/pkg/modules/salesfinance"
	"backend/pkg/modules/salesfinance/models"
	"backend/pkg/registry"
)

const maxFinanceUsers = 20

type salesFinanceUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewSalesFinanceUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) salesfinance.SalesFinanceUseCase {
	return &salesFinanceUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// settings reads userID's finance users and the options.
func (u *salesFinanceUseCase) settings(ctx context.Context, userID string) (*models.SalesFinanceSettings, error) {
	finance, err := u.repositoryRegistry.AppUserRepository.ListActiveByRole(ctx, role.Finance)
	if err != nil {
		return nil, err
	}
	links, err := u.repositoryRegistry.SalesFinanceRepository.ListBySales(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := &models.SalesFinanceSettings{Options: []models.SalesFinanceUser{}, FinanceUserIDs: []string{}}
	for _, f := range finance {
		if f.ID != userID {
			res.Options = append(res.Options, models.SalesFinanceUser{ID: f.ID, Name: f.Name})
		}
	}
	for _, l := range links {
		if usable(l.FinanceUser) {
			res.FinanceUserIDs = append(res.FinanceUserIDs, l.FinanceUserID)
			if l.IsDefault {
				res.DefaultFinanceUserID = l.FinanceUserID
			}
		}
	}
	return res, nil
}

// usable: an active user who still holds the finance role.
func usable(user *domain.AppUser) bool {
	return user != nil && user.Status && slices.Contains(user.Roles, role.Finance)
}
