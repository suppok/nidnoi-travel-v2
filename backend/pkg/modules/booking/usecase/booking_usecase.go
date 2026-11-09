package usecase

import (
	"backend/pkg/modules/booking"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type bookingUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewBookingUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) booking.BookingUseCase {
	return &bookingUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}
