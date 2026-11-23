package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/constants/errorcode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/linebooking"
	"backend/pkg/modules/linebooking/models"
	"backend/pkg/modules/media"
	mediamodels "backend/pkg/modules/media/models"
)

type LineBookingAPIHandler struct {
	router             fiber.Router
	apiHandler         *api.ApiHandler
	lineBookingUseCase linebooking.LineBookingUseCase
	mediaUseCase       media.MediaUseCase
	maxUploadBytes     int64
}

func NewLineBookingAPIHandler(
	router fiber.Router,
	apiHandler *api.ApiHandler,
	lineBookingUseCase linebooking.LineBookingUseCase,
	mediaUseCase media.MediaUseCase,
	maxUploadMB int,
) *LineBookingAPIHandler {
	return &LineBookingAPIHandler{
		router:             router,
		apiHandler:         apiHandler,
		lineBookingUseCase: lineBookingUseCase,
		mediaUseCase:       mediaUseCase,
		maxUploadBytes:     int64(maxUploadMB) * 1024 * 1024,
	}
}

func (h *LineBookingAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Booking entry (backoffice "บันทึกใบจอง LINE") — sales enter and reopen bookings. ---
	// Reopening isn't owner-scoped: sales share booking numbers.
	entry := []fiber.Handler{jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingCreate)}
	r.Post("/line-bookings", append(entry, h.LineBookingCreate)...)
	r.Post("/media/upload/line-booking-slip", append(entry, h.LineBookingSlipUpload)...)
	r.Post("/line-bookings/editable/list", append(entry, h.LineBookingEditableList)...)
	r.Put("/line-bookings/editable", append(entry, h.LineBookingUpdateDeal)...)
	r.Get("/line-bookings/editable/:bookingNo/events", append(entry, h.LineBookingEditableEventList)...)
	r.Get("/line-bookings/editable/:bookingNo", append(entry, h.LineBookingEditableGet)...)

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Put("/line-bookings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdate)
	r.Post("/line-bookings/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingList), h.LineBookingList)
	r.Put("/line-bookings/line-contact", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdateLineContact)
	r.Put("/line-bookings/confirm", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingConfirm)
	r.Put("/line-bookings/confirm-balance", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingConfirmBalance)
	r.Post("/line-bookings/comments", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingComment)
	r.Post("/line-booking-notifications/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingList), h.LineBookingNotificationList)
	r.Put("/line-booking-notifications/read", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingList), h.LineBookingNotificationRead)
	r.Put("/line-bookings/cancel", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingCancel)
	r.Put("/line-bookings/documents", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingDocument), h.LineBookingDocuments)
	r.Put("/line-bookings/documents/approve", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingDocumentsApprove)
	r.Put("/line-bookings/documents/reject", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingDocumentsReject)
	r.Post("/media/upload/line-booking", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingDocument), h.LineBookingDocumentUpload)
	r.Put("/line-bookings/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingStatus), h.LineBookingUpdateStatus)
	r.Put("/line-bookings/wholesale-email/resend", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingWholesaleEmail), h.LineBookingWholesaleEmailResend)
	r.Get("/line-bookings/:id/wholesale-email", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingWholesaleEmailGet)
	r.Get("/line-bookings/:id/mentionable", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingMentionable)
	r.Get("/line-bookings/:id/events", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingEventList)
	r.Get("/line-bookings/:id/finance-users", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingFinanceUserList)
	r.Get("/line-bookings/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingGet)
}

// LineBookingEditableGet godoc
// @Summary     Reopen a LINE booking by booking no
// @Description Only a paid booking (state 20 or 25), also while its documents wait for finance.
// @Description Any staff user with line-booking:create may open it (not limited to their own bookings).
// @Tags        line-bookings
// @Produce     json
// @Security    BearerAuth
// @Param       bookingNo path string true "Booking no"
// @Success     200 {object} models.LineBookingEditableGetResponse
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_NOT_EDITABLE"
// @Router      /line-bookings/editable/{bookingNo} [get]
func (h *LineBookingAPIHandler) LineBookingEditableGet(c *fiber.Ctx) error {
	request := &models.LineBookingEditableGetRequest{BookingNo: c.Params("bookingNo")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.EditableGet(ctx, request)
	})
}

// LineBookingEditableList godoc
// @Summary     LINE bookings of one customer chat that can be reopened
// @Description The chat's active paid bookings (state 20 or 25), newest confirmation first, at most 50.
// @Description Any staff user with line-booking:create may list them (sales share bookings).
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingEditableListRequest true "Customer chat"
// @Success     200 {object} models.LineBookingEditableListResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Router      /line-bookings/editable/list [post]
func (h *LineBookingAPIHandler) LineBookingEditableList(c *fiber.Ctx) error {
	request := &models.LineBookingEditableListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.EditableList(ctx, request)
	})
}

// LineBookingEditableEventList godoc
// @Summary     History of a reopened LINE booking
// @Description Same items as GET /line-bookings/{id}/events, for a booking that can be reopened (paid).
// @Tags        line-bookings
// @Produce     json
// @Security    BearerAuth
// @Param       bookingNo path string true "Booking no"
// @Success     200 {object} models.LineBookingEventListResponse
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_NOT_EDITABLE"
// @Router      /line-bookings/editable/{bookingNo}/events [get]
func (h *LineBookingAPIHandler) LineBookingEditableEventList(c *fiber.Ctx) error {
	request := &models.LineBookingEditableGetRequest{BookingNo: c.Params("bookingNo")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.EditableEvents(ctx, request)
	})
}

// LineBookingUpdateDeal godoc
// @Summary     Edit a reopened LINE booking, or record its balance
// @Description The whole deal again (like POST /line-bookings; tour, period, first slip and sales owner
// @Description stay). The total is quoted again. balanceSlipImages (1-10) record the balance, which then
// @Description waits for the sales owner's PUT /line-bookings/confirm-balance.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingDealUpdateRequest true "Booking"
// @Success     200 {object} models.LineBookingDealUpdateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_NOT_EDITABLE, PERIOD_UNAVAILABLE"
// @Router      /line-bookings/editable [put]
func (h *LineBookingAPIHandler) LineBookingUpdateDeal(c *fiber.Ctx) error {
	request := &models.LineBookingDealUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.UpdateDeal(ctx, request)
	})
}

// LineBookingConfirmBalance godoc
// @Summary     Confirm the balance of a deposit LINE booking
// @Description The sales owner (or an admin) checked the balance slip recorded on the booking entry page.
// @Description The state becomes 25 (paid in full) and the booking goes back to its finance user for the RE.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingConfirmBalanceRequest true "Booking"
// @Success     200 {object} models.LineBookingConfirmBalanceResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED (not the sales owner)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_BALANCE_NOT_PENDING"
// @Router      /line-bookings/confirm-balance [put]
func (h *LineBookingAPIHandler) LineBookingConfirmBalance(c *fiber.Ctx) error {
	request := &models.LineBookingConfirmBalanceRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.ConfirmBalance(ctx, request)
	})
}

// LineBookingCreate godoc
// @Summary     Create a LINE booking
// @Description Sales staff enter a booking with its payment slips on the backoffice booking entry page,
// @Description with the deal from the "สรุปยอดโอน" summary. Unit prices default to the period's; the total
// @Description includes discount, charges and card fee.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingCreateRequest true "Booking"
// @Success     200 {object} models.LineBookingCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND (tour)"
// @Failure     409 {object} map[string]any "PERIOD_UNAVAILABLE"
// @Router      /line-bookings [post]
func (h *LineBookingAPIHandler) LineBookingCreate(c *fiber.Ctx) error {
	request := &models.LineBookingCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Create(ctx, request)
	})
}

// LineBookingSlipUpload godoc
// @Summary     Upload a payment slip image for a LINE booking
// @Description multipart/form-data with "file"; the purpose is booking-slip (fixed by this endpoint).
// @Tags        line-bookings
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "Image file"
// @Success     200 {object} mediamodels.MediaUploadResponse
// @Failure     400 {object} map[string]any "VALIDATION, MEDIA_TYPE_NOT_ALLOWED, IMAGE_TOO_SMALL"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     413 {object} map[string]any "MEDIA_TOO_LARGE"
// @Failure     503 {object} map[string]any "STORAGE_UNAVAILABLE"
// @Router      /media/upload/line-booking-slip [post]
func (h *LineBookingAPIHandler) LineBookingSlipUpload(c *fiber.Ctx) error {
	request := &mediamodels.MediaUploadRequest{Purpose: mediamodels.PurposeBookingSlip}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		if err := h.readUpload(c, request); err != nil {
			return nil, err
		}
		return h.mediaUseCase.Upload(ctx, request)
	})
}

// readUpload reads the multipart "file" field into request (no file = an empty upload, which the
// media usecase rejects).
func (h *LineBookingAPIHandler) readUpload(c *fiber.Ctx, request *mediamodels.MediaUploadRequest) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return nil
	}
	if fh.Size > h.maxUploadBytes {
		return cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
	}
	f, err := fh.Open()
	if err != nil {
		return cerrors.Wrap(err, "open upload")
	}
	defer f.Close()
	if request.Data, err = io.ReadAll(io.LimitReader(f, h.maxUploadBytes+1)); err != nil {
		return cerrors.Wrap(err, "read upload")
	}
	request.FileName = fh.Filename
	return nil
}

// LineBookingDocumentUpload godoc
// @Summary     Upload a finance document PDF of a LINE booking
// @Description multipart/form-data with "file" (PDF). Stored under bookings/documents/; send the URL with
// @Description PUT /line-bookings/documents.
// @Tags        line-bookings
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "PDF (max media.maxUploadMB)"
// @Success     200 {object} mediamodels.MediaUploadResponse
// @Failure     400 {object} map[string]any "VALIDATION, MEDIA_TYPE_NOT_ALLOWED"
// @Failure     413 {object} map[string]any "MEDIA_TOO_LARGE"
// @Failure     503 {object} map[string]any "STORAGE_UNAVAILABLE"
// @Router      /media/upload/line-booking [post]
func (h *LineBookingAPIHandler) LineBookingDocumentUpload(c *fiber.Ctx) error {
	request := &mediamodels.MediaUploadRequest{Purpose: mediamodels.PurposeBookingDocument}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		if err := h.readUpload(c, request); err != nil {
			return nil, err
		}
		return h.mediaUseCase.Upload(ctx, request)
	})
}

// LineBookingDocumentsApprove godoc
// @Summary     Approve the finance documents and send them to the customer
// @Description The sales owner (or an admin). When the booking is linked to a LINE chat whose customer
// @Description follows the OA, a Flex card with a button per PDF is pushed there first (one message of
// @Description the OA quota); otherwise the documents are only approved.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingDocumentsApproveRequest true "Booking"
// @Success     200 {object} models.LineBookingDocumentsApproveResponse
// @Failure     403 {object} map[string]any "PERMISSION_DENIED (not the sales owner)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_DOCUMENTS_NOT_READY"
// @Failure     429 {object} map[string]any "LINE_QUOTA_EXCEEDED"
// @Failure     502 {object} map[string]any "LINE_PUSH_FAILED"
// @Failure     503 {object} map[string]any "LINE_NOT_CONFIGURED"
// @Router      /line-bookings/documents/approve [put]
func (h *LineBookingAPIHandler) LineBookingDocumentsApprove(c *fiber.Ctx) error {
	request := &models.LineBookingDocumentsApproveRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.DocumentsApprove(ctx, request)
	})
}

// LineBookingDocumentsReject godoc
// @Summary     Send the finance documents back to finance with a note
// @Description The sales owner (or an admin). The booking waits for its finance user again.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingDocumentsRejectRequest true "Booking and note"
// @Success     200 {object} models.LineBookingDocumentsRejectResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED (not the sales owner)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_DOCUMENTS_NOT_READY"
// @Router      /line-bookings/documents/reject [put]
func (h *LineBookingAPIHandler) LineBookingDocumentsReject(c *fiber.Ctx) error {
	request := &models.LineBookingDocumentsRejectRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.DocumentsReject(ctx, request)
	})
}

// LineBookingDocuments godoc
// @Summary     Send the finance documents of a confirmed LINE booking
// @Description Its finance user (or an admin). A deposit booking needs aiFile (ใบรับเงินมัดจำ) + ivFile
// @Description (ใบแจ้งหนี้); a full payment needs reFile (ใบเสร็จรับเงิน). The first send sets documentsAt,
// @Description which alerts the sales owner; later sends replace the files.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingDocumentsRequest true "Documents"
// @Success     200 {object} models.LineBookingDocumentsResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED (not its finance user)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/documents [put]
func (h *LineBookingAPIHandler) LineBookingDocuments(c *fiber.Ctx) error {
	request := &models.LineBookingDocumentsRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Documents(ctx, request)
	})
}

// LineBookingList godoc
// @Summary  List LINE bookings, newest first
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingListRequest true "Filter"
// @Success  200 {object} models.LineBookingListResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Router   /line-bookings/list [post]
func (h *LineBookingAPIHandler) LineBookingList(c *fiber.Ctx) error {
	request := &models.LineBookingListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.List(ctx, request)
	})
}

// LineBookingComment godoc
// @Summary     Comment on a LINE booking
// @Description Adds a comment to the booking's history. Anyone who can see the booking may comment. The sales
// @Description owner and the finance user get a "comment" notification, users in mentionUserIds (from
// @Description GET /line-bookings/{id}/mentionable) a "mention" one; never the author.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingCommentRequest true "Comment"
// @Success     200 {object} models.LineBookingCommentResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/comments [post]
func (h *LineBookingAPIHandler) LineBookingComment(c *fiber.Ctx) error {
	request := &models.LineBookingCommentRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Comment(ctx, request)
	})
}

// LineBookingMentionable godoc
// @Summary  Users who may be @mentioned on a LINE booking
// @Description The people who can see it: its sales owner, its finance user and the admins.
// @Tags     line-bookings
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "LINE Booking ID"
// @Success  200 {object} models.LineBookingMentionableResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /line-bookings/{id}/mentionable [get]
func (h *LineBookingAPIHandler) LineBookingMentionable(c *fiber.Ctx) error {
	request := &models.LineBookingMentionableRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Mentionable(ctx, request)
	})
}

// LineBookingNotificationList godoc
// @Summary  The caller's LINE booking comment notifications
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingNotificationListRequest true "Filter"
// @Success  200 {object} models.LineBookingNotificationListResponse
// @Router   /line-booking-notifications/list [post]
func (h *LineBookingAPIHandler) LineBookingNotificationList(c *fiber.Ctx) error {
	request := &models.LineBookingNotificationListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Notifications(ctx, request)
	})
}

// LineBookingNotificationRead godoc
// @Summary  Mark the caller's notifications on a LINE booking read
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingNotificationReadRequest true "Booking"
// @Success  200 {object} map[string]any
// @Failure  400 {object} map[string]any "VALIDATION"
// @Router   /line-booking-notifications/read [put]
func (h *LineBookingAPIHandler) LineBookingNotificationRead(c *fiber.Ctx) error {
	request := &models.LineBookingNotificationReadRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return nil, h.lineBookingUseCase.NotificationsRead(ctx, request)
	})
}

// LineBookingEventList godoc
// @Summary  A LINE booking's history
// @Description Every action on the booking, oldest first: created, confirmed, documents, deal edits (with
// @Description the changed fields), balance, LINE chat, staff note, delete/restore. Same visibility as GET /line-bookings/{id}.
// @Tags     line-bookings
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "LINE Booking ID"
// @Success  200 {object} models.LineBookingEventListResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /line-bookings/{id}/events [get]
func (h *LineBookingAPIHandler) LineBookingEventList(c *fiber.Ctx) error {
	request := &models.LineBookingEventListRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Events(ctx, request)
	})
}

// LineBookingGet godoc
// @Summary  Get a LINE booking
// @Tags     line-bookings
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "LINE Booking ID"
// @Success  200 {object} models.LineBookingGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /line-bookings/{id} [get]
func (h *LineBookingAPIHandler) LineBookingGet(c *fiber.Ctx) error {
	request := &models.LineBookingGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Get(ctx, request)
	})
}

// LineBookingUpdate godoc
// @Summary  Update a LINE booking's staff note
// @Description The state changes only through PUT /line-bookings/confirm and /cancel.
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingUpdateRequest true "Note"
// @Success  200 {object} models.LineBookingUpdateResponse
// @Failure  400 {object} map[string]any "VALIDATION"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /line-bookings [put]
func (h *LineBookingAPIHandler) LineBookingUpdate(c *fiber.Ctx) error {
	request := &models.LineBookingUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Update(ctx, request)
	})
}

// LineBookingConfirm godoc
// @Summary     Confirm a new LINE booking and send it to a finance user
// @Description The sales owner (or an admin) checked the payment. The state becomes 20 (deposit paid) or
// @Description 25 (paid in full) by the payment plan; financeUserId must be one of
// @Description GET /line-bookings/{id}/finance-users.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingConfirmRequest true "Booking and finance user"
// @Success     200 {object} models.LineBookingConfirmResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED (not the sales owner)"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Failure     409 {object} map[string]any "LINE_BOOKING_NOT_NEW"
// @Router      /line-bookings/confirm [put]
func (h *LineBookingAPIHandler) LineBookingConfirm(c *fiber.Ctx) error {
	request := &models.LineBookingConfirmRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Confirm(ctx, request)
	})
}

// LineBookingCancel godoc
// @Summary  Cancel a new LINE booking
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingCancelRequest true "Booking"
// @Success  200 {object} models.LineBookingCancelResponse
// @Failure  403 {object} map[string]any "PERMISSION_DENIED (not the sales owner)"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Failure  409 {object} map[string]any "LINE_BOOKING_NOT_NEW"
// @Router   /line-bookings/cancel [put]
func (h *LineBookingAPIHandler) LineBookingCancel(c *fiber.Ctx) error {
	request := &models.LineBookingCancelRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Cancel(ctx, request)
	})
}

// LineBookingFinanceUserList godoc
// @Summary     Finance users a LINE booking can be sent to on confirm
// @Description The sales owner's active finance users and their default (every active finance user
// @Description when the booking has no owner).
// @Tags        line-bookings
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "LINE booking ID"
// @Success     200 {object} models.LineBookingFinanceUserListResponse
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/{id}/finance-users [get]
func (h *LineBookingAPIHandler) LineBookingFinanceUserList(c *fiber.Ctx) error {
	request := &models.LineBookingFinanceUserListRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.FinanceUserList(ctx, request)
	})
}

// LineBookingUpdateLineContact godoc
// @Summary     Link a LINE booking to the customer's LINE OA chat
// @Description lineContactId comes from POST /line-contacts/list; null unlinks. The receipt is
// @Description pushed to this chat later.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingUpdateLineContactRequest true "Booking and contact"
// @Success     200 {object} models.LineBookingUpdateLineContactResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/line-contact [put]
func (h *LineBookingAPIHandler) LineBookingUpdateLineContact(c *fiber.Ctx) error {
	request := &models.LineBookingUpdateLineContactRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.UpdateLineContact(ctx, request)
	})
}

// LineBookingUpdateStatus godoc
// @Summary  Delete or restore a LINE booking
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingUpdateStatusRequest true "Status"
// @Success  200 {object} map[string]any "data: null"
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /line-bookings/status [put]
func (h *LineBookingAPIHandler) LineBookingUpdateStatus(c *fiber.Ctx) error {
	request := &models.LineBookingUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.UpdateStatus(ctx, request)
	})
}

// LineBookingWholesaleEmailGet godoc
// @Summary     The booking email to the wholesale
// @Description State of the email queued to the tour's wholesale when the booking was saved: 10 pending
// @Description (queued or waiting to retry), 20 sent, 30 failed. email is null when none was queued.
// @Description Same visibility as GET /line-bookings/{id}.
// @Tags        line-bookings
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "LINE Booking ID"
// @Success     200 {object} models.LineBookingWholesaleEmailGetResponse
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/{id}/wholesale-email [get]
func (h *LineBookingAPIHandler) LineBookingWholesaleEmailGet(c *fiber.Ctx) error {
	request := &models.LineBookingWholesaleEmailGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.WholesaleEmailGet(ctx, request)
	})
}

// LineBookingWholesaleEmailResend godoc
// @Summary     Resend the booking email to the wholesale
// @Description Sales owner or admin. Re-reads the wholesale's current addresses and re-renders the
// @Description email, then queues it again (pending, attempts 0, a new idempotency key). Works even
// @Description when the wholesale had auto email off when the booking was saved. VALIDATION on id
// @Description when the booking is cancelled or the wholesale has no email address.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.LineBookingWholesaleEmailResendRequest true "Booking"
// @Success     200 {object} models.LineBookingWholesaleEmailResendResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     403 {object} map[string]any "PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND"
// @Router      /line-bookings/wholesale-email/resend [put]
func (h *LineBookingAPIHandler) LineBookingWholesaleEmailResend(c *fiber.Ctx) error {
	request := &models.LineBookingWholesaleEmailResendRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.WholesaleEmailResend(ctx, request)
	})
}
