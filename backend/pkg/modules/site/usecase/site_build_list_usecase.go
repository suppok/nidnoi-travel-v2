package usecase

import (
	"context"

	"backend/pkg/constants/sitebuild"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/site/models"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func (u *siteUseCase) BuildListValidate(request *models.SiteBuildListRequest) error {
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	triggers := make([]any, len(sitebuild.Triggers))
	for i, t := range sitebuild.Triggers {
		triggers[i] = t
	}
	return validation.ValidateStruct(request,
		validation.Field(&request.Trigger, validation.In(triggers...)),
		validation.Field(&request.Page, validation.Min(1)),
		validation.Field(&request.PageSize, validation.Min(1), validation.Max(maxPageSize)),
	)
}

// BuildList returns the backend-triggered customer-site builds, newest first.
func (u *siteUseCase) BuildList(ctx context.Context, request *models.SiteBuildListRequest) (*models.SiteBuildListResponse, error) {
	if err := u.BuildListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	builds, total, err := u.repositoryRegistry.SiteBuildRepository.List(ctx, repomodels.SiteBuildListCriteria{
		Trigger:  request.Trigger,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list site builds")
	}
	items := make([]models.SiteBuildItem, len(builds))
	for i := range builds {
		items[i] = models.NewSiteBuildItem(&builds[i])
	}
	return &models.SiteBuildListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
