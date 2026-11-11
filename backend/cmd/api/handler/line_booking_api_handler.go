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

func (h *LineBookingAPIHandler) Init(jwtAuth middleware.Auth, liffAuth middleware.LIFFAuth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- LIFF (LINE app) — Authorization is the LIFF access token of a linked staff user. ---
	liff := []fiber.Handler{liffAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingCreate)}
	r.Post("/line-bookings", append(liff, h.LineBookingCreate)...)
	r.Post("/media/upload/liff", append(liff, h.LIFFMediaUpload)...)
	r.Get("/liff/sales-users", append(liff, h.LIFFSalesUserList)...)
	r.Get("/liff/booking-sources", append(liff, h.LIFFBookingSourceList)...)

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Put("/line-bookings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdate)
	r.Post("/line-bookings/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingList), h.LineBookingList)
	r.Put("/line-bookings/line-contact", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdateLineContact)
	r.Put("/line-bookings/confirm", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingConfirm)
	r.Put("/line-bookings/cancel", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingCancel)
	r.Put("/line-bookings/documents", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingDocument), h.LineBookingDocuments)
	r.Put("/line-bookings/documents/approve", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingDocumentsApprove)
	r.Put("/line-bookings/documents/reject", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingDocumentsReject)
	r.Post("/media/upload/line-booking", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingDocument), h.LineBookingDocumentUpload)
	r.Put("/line-bookings/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingStatus), h.LineBookingUpdateStatus)
	r.Get("/line-bookings/:id/finance-users", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingFinanceUserList)
	r.Get("/line-bookings/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingGet)
}

// LineBookingCreate godoc
// @Summary     Create a LINE booking from LIFF
// @Description Sales staff submit a booking with payment slip from the LIFF app, with the deal from the
// @Description "สรุปยอดโอน" summary. Unit prices default to the period's; the total includes discount,
// @Description charges and card fee. Authorization: Bearer <LIFF access token> of a LINE account linked
// @Description to a staff user with line-booking:create.
// @Tags        line-bookings
// @Accept      json
// @Produce     json
// @Param       body body models.LineBookingCreateRequest true "Booking"
// @Success     200 {object} models.LineBookingCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     401 {object} map[string]any "INVALID_LINE_TOKEN"
// @Failure     403 {object} map[string]any "LINE_NOT_LINKED, PERMISSION_DENIED"
// @Failure     404 {object} map[string]any "NOT_FOUND (tour)"
// @Failure     409 {object} map[string]any "PERIOD_UNAVAILABLE"
// @Failure     503 {object} map[string]any "LINE_NOT_CONFIGURED"
// @Router      /line-bookings [post]
func (h *LineBookingAPIHandler) LineBookingCreate(c *fiber.Ctx) error {
	request := &models.LineBookingCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.Create(ctx, request)
	})
}

// LIFFMediaUpload godoc
// @Summary     Upload a payment slip image from LIFF
// @Description multipart/form-data with "file" and "purpose" (booking-slip).
// @Description Authorization: Bearer <LIFF access token> of a linked staff user with line-booking:create.
// @Tags        line-bookings
// @Accept      multipart/form-data
// @Produce     json
// @Param       file            formData file   true "Image file"
// @Param       purpose         formData string true "Purpose (booking-slip)"
// @Success     200 {object} mediamodels.MediaUploadResponse
// @Failure     400 {object} map[string]any "VALIDATION, MEDIA_TYPE_NOT_ALLOWED, IMAGE_TOO_SMALL"
// @Failure     401 {object} map[string]any "INVALID_LINE_TOKEN"
// @Failure     403 {object} map[string]any "LINE_NOT_LINKED, PERMISSION_DENIED"
// @Failure     413 {object} map[string]any "MEDIA_TOO_LARGE"
// @Failure     503 {object} map[string]any "STORAGE_UNAVAILABLE, LINE_NOT_CONFIGURED"
// @Router      /media/upload/liff [post]
func (h *LineBookingAPIHandler) LIFFMediaUpload(c *fiber.Ctx) error {
	request := &mediamodels.MediaUploadRequest{Purpose: c.FormValue("purpose")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		// The LIFF app may only upload payment slips.
		if request.Purpose != mediamodels.PurposeBookingSlip {
			return nil, cerrors.New(http.StatusBadRequest, errorcode.MediaTypeNotAllowed)
		}
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

// LIFFSalesUserList godoc
// @Summary     Sales owner options for the LIFF booking form
// @Description Active users with the sales role, by name; defaultId is the caller when they are sales.
// @Description Authorization: Bearer <LIFF access token> of a linked staff user with line-booking:create.
// @Tags        liff
// @Produce     json
// @Success     200 {object} models.LineBookingSalesUserListResponse
// @Failure     403 {object} map[string]any "LINE_NOT_LINKED, PERMISSION_DENIED"
// @Router      /liff/sales-users [get]
func (h *LineBookingAPIHandler) LIFFSalesUserList(c *fiber.Ctx) error {
	request := &models.LineBookingSalesUserListRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.SalesUserList(ctx, request)
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

// LIFFBookingSourceList godoc
// @Summary     Booking source options for the LIFF booking form
// @Description Active booking sources in their sort order.
// @Description Authorization: Bearer <LIFF access token> of a linked staff user with line-booking:create.
// @Tags        liff
// @Produce     json
// @Success     200 {object} models.LineBookingSourceListResponse
// @Failure     403 {object} map[string]any "LINE_NOT_LINKED, PERMISSION_DENIED"
// @Router      /liff/booking-sources [get]
func (h *LineBookingAPIHandler) LIFFBookingSourceList(c *fiber.Ctx) error {
	request := &models.LineBookingSourceListRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.lineBookingUseCase.SourceList(ctx, request)
	})
}
