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

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Put("/line-bookings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdate)
	r.Post("/line-bookings/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingList), h.LineBookingList)
	r.Put("/line-bookings/line-contact", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingUpdate), h.LineBookingUpdateLineContact)
	r.Put("/line-bookings/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingStatus), h.LineBookingUpdateStatus)
	r.Get("/line-bookings/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.LineBookingGet), h.LineBookingGet)
}

// LineBookingCreate godoc
// @Summary     Create a LINE booking from LIFF
// @Description Sales staff submit a booking with payment slip from the LIFF app. Prices are quoted
// @Description from the period. Authorization: Bearer <LIFF access token> of a LINE account linked
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

		if fh, err := c.FormFile("file"); err == nil {
			if fh.Size > h.maxUploadBytes {
				return nil, cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
			}
			f, err := fh.Open()
			if err != nil {
				return nil, cerrors.Wrap(err, "open upload")
			}
			defer f.Close()
			if request.Data, err = io.ReadAll(io.LimitReader(f, h.maxUploadBytes+1)); err != nil {
				return nil, cerrors.Wrap(err, "read upload")
			}
			request.FileName = fh.Filename
		}
		return h.mediaUseCase.Upload(ctx, request)
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
// @Summary  Update a LINE booking's state and staff note
// @Tags     line-bookings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.LineBookingUpdateRequest true "State and note"
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
