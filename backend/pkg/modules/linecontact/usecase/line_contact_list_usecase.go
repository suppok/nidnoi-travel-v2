package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/linecontact/models"
)

func (u *lineContactUseCase) ListValidate(request *models.LineContactListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *lineContactUseCase) List(ctx context.Context, request *models.LineContactListRequest) (*models.LineContactListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	items, total, err := u.repositoryRegistry.LineContactRepository.List(ctx, repomodels.LineContactListCriteria{
		Keyword:   request.Keyword,
		Following: request.Following,
		Page:      request.Page,
		PageSize:  request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list line contacts")
	}
	result := make([]models.LineContactItem, len(items))
	for i := range items {
		result[i] = models.NewLineContactItem(&items[i])
	}
	return &models.LineContactListResponse{Items: result, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
