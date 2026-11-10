package usecase

import (
	"backend/pkg/modules/linecontact"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type lineContactUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewLineContactUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) linecontact.LineContactUseCase {
	return &lineContactUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}
