package usecase

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/linebookingstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

// documentPrefix is where POST /media/upload/line-booking stores finance PDFs (purpose booking-document).
const documentPrefix = "bookings/documents/"

// documentURL accepts "" or a PDF uploaded through POST /media/upload/line-booking.
func (u *lineBookingUseCase) documentURL() validation.Rule {
	return validation.By(func(value any) error {
		s, _ := value.(string)
		if s == "" {
			return nil
		}
		base := u.shareRegistry.ClientRegistry.ObjectStorage.PublicBaseURL()
		if base == "" {
			return errors.New("file storage is not configured")
		}
		rest, ok := strings.CutPrefix(s, base+"/"+documentPrefix)
		if !ok || rest == "" || strings.ContainsAny(rest, "?#") || strings.Contains(rest, "..") || !strings.HasSuffix(strings.ToLower(rest), ".pdf") {
			return errors.New("must be a PDF uploaded through /media/upload/line-booking")
		}
		return nil
	})
}

func (u *lineBookingUseCase) DocumentsValidate(request *models.LineBookingDocumentsRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.AIFile, u.documentURL()),
		validation.Field(&request.IVFile, u.documentURL()),
		validation.Field(&request.REFile, u.documentURL()),
	)
}

// documentsForPlan checks the files match the payment plan: a deposit needs AI + IV, a full
// payment needs RE, and nothing else.
func documentsForPlan(b *domain.LineBooking, r *models.LineBookingDocumentsRequest) error {
	errs := validation.Errors{}
	if b.IsDeposit {
		if r.AIFile == "" {
			errs["aiFile"] = errors.New("required for a deposit")
		}
		if r.IVFile == "" {
			errs["ivFile"] = errors.New("required for a deposit")
		}
		if r.REFile != "" {
			errs["reFile"] = errors.New("must be empty for a deposit")
		}
	} else {
		if r.REFile == "" {
			errs["reFile"] = errors.New("required for a full payment")
		}
		if r.AIFile != "" {
			errs["aiFile"] = errors.New("must be empty for a full payment")
		}
		if r.IVFile != "" {
			errs["ivFile"] = errors.New("must be empty for a full payment")
		}
	}
	if len(errs) > 0 {
		return cerrors.Validation(errs)
	}
	return nil
}

// Documents saves the finance documents of a confirmed booking. Only its finance user (or an admin)
// may; the first save sets documents_at, which alerts the sales owner. Later saves replace files.
func (u *lineBookingUseCase) Documents(ctx context.Context, request *models.LineBookingDocumentsRequest) (*models.LineBookingDocumentsResponse, error) {
	request.AIFile = strings.TrimSpace(request.AIFile)
	request.IVFile = strings.TrimSpace(request.IVFile)
	request.REFile = strings.TrimSpace(request.REFile)
	if err := u.DocumentsValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		info := request.UserRequestInfo
		if target == nil || !target.Status || !visibleTo(target, info) {
			return cerrors.NotFound()
		}
		if !sendsDocuments(target, info) {
			return cerrors.PermissionDenied()
		}
		if !slices.Contains(linebookingstate.Paid, target.State) {
			return cerrors.FieldError("id", "the booking must be confirmed first")
		}
		if target.DocumentsApprovedAt.Valid {
			return cerrors.FieldError("id", "the documents were already approved by sales")
		}
		if err := documentsForPlan(target, request); err != nil {
			return err
		}

		now := time.Now()
		target.AIFile, target.IVFile, target.REFile = request.AIFile, request.IVFile, request.REFile
		columns := []string{"ai_file", "iv_file", "re_file", "update_date", "update_by"}
		if !target.DocumentsAt.Valid {
			target.DocumentsAt = null.TimeFrom(now)
			target.DocumentsBy = null.StringFrom(info.UserID)
			columns = append(columns, "documents_at", "documents_by")
		}
		target.UpdateDate = now
		target.UpdateBy = info.UserID
		if _, err = repo.Update(ctx, target, columns...); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "save line booking documents")
	}
	return &models.LineBookingDocumentsResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}
