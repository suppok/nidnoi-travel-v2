package usecase

import (
	"context"
	"time"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/modules/tourprogram/models"
)

func (u *tourProgramUseCase) PublicGet(ctx context.Context, request *models.TourProgramPublicGetRequest) (*models.TourProgramPublicGetResponse, error) {
	if !publicCodePattern.MatchString(request.Code) {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.TourProgramRepository.FindPublicByCode(ctx, request.Code)
	if err != nil {
		return nil, cerrors.Wrap(err, "get public tour program")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}

	periods, err := u.repositoryRegistry.TourProgramPeriodRepository.ListPublicByTours(ctx, []string{found.ID}, utils.TodayInThailand(time.Now()))
	if err != nil {
		return nil, cerrors.Wrap(err, "list public periods")
	}
	return &models.TourProgramPublicGetResponse{TourProgramPublicDetail: models.NewTourProgramPublicDetail(found, periods)}, nil
}
