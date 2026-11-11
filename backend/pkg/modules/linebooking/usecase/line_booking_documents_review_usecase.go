package usecase

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
	"backend/pkg/modules/shared/client/line"
	"backend/pkg/repository"
)

const maxFixNoteRunes = 1000

// retryNamespace scopes the LINE retry keys of documents pushes (uuid v5 of booking + documents time).
var retryNamespace = uuid.MustParse("6f1d7c2e-3b7a-4c1e-9a52-1f0e0b6d4a10")

// loadForReview loads a booking whose documents wait for the sales owner's review: NOT_FOUND when
// hidden or deleted, PERMISSION_DENIED for anyone but the owner (or an admin),
// LINE_BOOKING_DOCUMENTS_NOT_READY when finance hasn't sent them or they were already approved.
func loadForReview(ctx context.Context, repo repository.LineBookingRepository, id string, info *coremodels.UserRequestInfo) (*domain.LineBooking, error) {
	b, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil || !b.Status || !visibleTo(b, info) {
		return nil, cerrors.NotFound()
	}
	if !ownsAsSales(b, info) {
		return nil, cerrors.PermissionDenied()
	}
	if !b.DocumentsAt.Valid || b.DocumentsApprovedAt.Valid {
		return nil, cerrors.New(http.StatusConflict, errorcode.LineBookingDocumentsNotReady)
	}
	return b, nil
}

// pushTarget is the chat the documents go to: the linked LINE contact, when the customer still
// follows the OA; nil = approve without sending.
func pushTarget(b *domain.LineBooking) *domain.LineContact {
	if c := b.LineContact; c != nil && c.Following {
		return c
	}
	return nil
}

func (u *lineBookingUseCase) DocumentsApproveValidate(request *models.LineBookingDocumentsApproveRequest) error {
	return validation.ValidateStruct(request, validation.Field(&request.ID, validation.Required, validation.UUID))
}

// DocumentsApprove is the sales owner's "ส่งให้ลูกค้า": the documents are right. When the booking
// is linked to a chat whose customer follows the OA, a card with a button per PDF is pushed there
// first; a failed push approves nothing.
func (u *lineBookingUseCase) DocumentsApprove(ctx context.Context, request *models.LineBookingDocumentsApproveRequest) (*models.LineBookingDocumentsApproveResponse, error) {
	if err := u.DocumentsApproveValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	info := request.UserRequestInfo

	b, err := loadForReview(ctx, u.repositoryRegistry.LineBookingRepository, request.ID, info)
	if err != nil {
		return nil, cerrors.Wrap(err, "approve line booking documents")
	}

	target := pushTarget(b)
	if target != nil {
		key := uuid.NewSHA1(retryNamespace, []byte(b.ID+"|"+b.DocumentsAt.Time.UTC().Format(time.RFC3339Nano))).String()
		if err := u.shareRegistry.ClientRegistry.LINE.PushMessage(ctx, target.LineUserID, documentsMessages(b), key); err != nil {
			switch {
			case errors.Is(err, line.ErrNotConfigured):
				return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
			case errors.Is(err, line.ErrPushQuota):
				return nil, cerrors.New(http.StatusTooManyRequests, errorcode.LINEQuotaExceeded)
			default:
				u.coreRegistry.Logger.Error("LINE documents push failed", zap.String("bookingNo", b.BookingNo), zap.Error(err))
				return nil, cerrors.New(http.StatusBadGateway, errorcode.LINEPushFailed)
			}
		}
	}

	var updated *domain.LineBooking
	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		current, err := loadForReview(ctx, repo, request.ID, info)
		if err != nil {
			return err
		}
		now := time.Now()
		current.DocumentsApprovedAt = null.TimeFrom(now)
		current.DocumentsApprovedBy = null.StringFrom(info.UserID)
		columns := []string{"documents_approved_at", "documents_approved_by", "update_date", "update_by"}
		if target != nil {
			current.CustomerSentAt = null.TimeFrom(now)
			current.CustomerSentContactID = null.StringFrom(target.ID)
			columns = append(columns, "customer_sent_at", "customer_sent_contact_id")
		}
		current.UpdateDate = now
		current.UpdateBy = info.UserID
		if _, err = repo.Update(ctx, current, columns...); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, current.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "approve line booking documents")
	}
	return &models.LineBookingDocumentsApproveResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}

func (u *lineBookingUseCase) DocumentsRejectValidate(request *models.LineBookingDocumentsRejectRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Note, validation.Required, validation.RuneLength(1, maxFixNoteRunes)),
	)
}

// DocumentsReject is the sales owner's "แจ้งแก้ไข": the booking goes back to its finance user with
// the note (documents_at cleared, so it is waiting again); the files stay for finance to replace.
func (u *lineBookingUseCase) DocumentsReject(ctx context.Context, request *models.LineBookingDocumentsRejectRequest) (*models.LineBookingDocumentsRejectResponse, error) {
	request.Note = strings.TrimSpace(request.Note)
	if err := u.DocumentsRejectValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	info := request.UserRequestInfo

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		b, err := loadForReview(ctx, repo, request.ID, info)
		if err != nil {
			return err
		}
		now := time.Now()
		b.FixNote = request.Note
		b.FixRequestedAt = null.TimeFrom(now)
		b.DocumentsAt = null.Time{}
		b.DocumentsBy = null.String{}
		b.UpdateDate = now
		b.UpdateBy = info.UserID
		if _, err = repo.Update(ctx, b, "fix_note", "fix_requested_at", "documents_at", "documents_by", "update_date", "update_by"); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, b.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "reject line booking documents")
	}
	return &models.LineBookingDocumentsRejectResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}
