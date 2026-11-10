package registry

import (
	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"backend/pkg/core/auth/jwt"
	"backend/pkg/core/auth/permission"
	"backend/pkg/core/config"
	"backend/pkg/core/database/query"
	"backend/pkg/modules/shared/client/deployhook"
	"backend/pkg/modules/shared/client/line"
	"backend/pkg/modules/shared/client/objectstorage"
	"backend/pkg/modules/shared/services/password"
	"backend/pkg/modules/shared/services/revalidation"
	"backend/pkg/repository"
)

// CoreRegistry holds infrastructure shared by every module.
type CoreRegistry struct {
	Config      *config.Config
	Logger      *zap.Logger
	AppDB       *bun.DB
	DBTx        query.DBTx
	JWT         jwt.Service
	Permissions permission.Checker
}

// RepositoryRegistry holds every repository, bound to the non-transactional DB.
// Built in pkg/app/initializer/repository_initializer.go.
type RepositoryRegistry struct {
	AppUserRepository          repository.AppUserRepository
	RefreshTokenRepository     repository.RefreshTokenRepository
	CountryRepository          repository.CountryRepository
	CityRepository             repository.CityRepository
	WholesaleRepository        repository.WholesaleRepository
	WholesaleContactRepository repository.WholesaleContactRepository
	AirlineRepository          repository.AirlineRepository

	TourProgramRepository        repository.TourProgramRepository
	TourProgramCountryRepository repository.TourProgramCountryRepository
	TourProgramCityRepository    repository.TourProgramCityRepository
	TourProgramImageRepository   repository.TourProgramImageRepository
	TourProgramPeriodRepository  repository.TourProgramPeriodRepository

	SiteBuildRepository repository.SiteBuildRepository

	BookingRepository     repository.BookingRepository
	LineBookingRepository repository.LineBookingRepository
	LineContactRepository repository.LineContactRepository
}

// ShareRegistry holds cross-module services and outbound clients.
type ShareRegistry struct {
	ServiceRegistry ServiceRegistry
	ClientRegistry  ClientRegistry
}

type ServiceRegistry struct {
	Password    password.Service
	Revalidator revalidation.Revalidator
}

type ClientRegistry struct {
	ObjectStorage objectstorage.ObjectStorage
	DeployHook    deployhook.DeployHook
	LINE          line.Client
}
