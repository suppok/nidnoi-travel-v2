package initializer

import (
	"backend/pkg/core/database/query"
	"backend/pkg/registry"
	"backend/pkg/repository"
)

// NewRepositoryRegistry registers every repository. Add new repositories here.
func NewRepositoryRegistry(db query.AppIDB) *registry.RepositoryRegistry {
	return &registry.RepositoryRegistry{
		AppUserRepository:          repository.NewAppUserRepository(db),
		RefreshTokenRepository:     repository.NewRefreshTokenRepository(db),
		CountryRepository:          repository.NewCountryRepository(db),
		CityRepository:             repository.NewCityRepository(db),
		WholesaleRepository:        repository.NewWholesaleRepository(db),
		WholesaleContactRepository: repository.NewWholesaleContactRepository(db),
		AirlineRepository:          repository.NewAirlineRepository(db),

		TourProgramRepository:        repository.NewTourProgramRepository(db),
		TourProgramCountryRepository: repository.NewTourProgramCountryRepository(db),
		TourProgramCityRepository:    repository.NewTourProgramCityRepository(db),
		TourProgramImageRepository:   repository.NewTourProgramImageRepository(db),
		TourProgramPeriodRepository:  repository.NewTourProgramPeriodRepository(db),
	}
}
