package usecase

import (
	"backend/pkg/modules/linebooking"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type lineBookingUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewLineBookingUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) linebooking.LineBookingUseCase {
	return &lineBookingUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}
