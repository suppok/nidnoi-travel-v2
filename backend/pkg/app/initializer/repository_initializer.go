package initializer

import (
	"backend/pkg/core/database/query"
	"backend/pkg/registry"
	"backend/pkg/repository"
)

// NewRepositoryRegistry registers every repository. Add new repositories here.
func NewRepositoryRegistry(db query.AppIDB) *registry.RepositoryRegistry {
	return &registry.RepositoryRegistry{
		AppUserRepository:      repository.NewAppUserRepository(db),
		RefreshTokenRepository: repository.NewRefreshTokenRepository(db),
	}
}
