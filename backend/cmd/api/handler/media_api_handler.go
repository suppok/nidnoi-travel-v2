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
	"backend/pkg/modules/media"
	"backend/pkg/modules/media/models"
)

type MediaAPIHandler struct {
	router       fiber.Router
	apiHandler   *api.ApiHandler
	mediaUseCase media.MediaUseCase
	maxBytes     int64
}

func NewMediaAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, mediaUseCase media.MediaUseCase, maxUploadMB int) *MediaAPIHandler {
	return &MediaAPIHandler{router: router, apiHandler: apiHandler, mediaUseCase: mediaUseCase, maxBytes: int64(maxUploadMB) * 1024 * 1024}
}

func (h *MediaAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	h.router.Post("/media/upload", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.MediaUpload), h.MediaUpload)
}

// MediaUpload godoc
// @Summary     Upload an image or document to R2
// @Description multipart/form-data with "file" and "purpose" (tour-banner | tour-card | tour-gallery | tour-pdf | tour-docx | country-hero | city-hero).
// @Description The type is detected from the content: images are JPEG/PNG/WebP with a minimum size per purpose, documents are PDF or DOCX.
// @Tags        media
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file    formData file   true "File (max media.maxUploadMB)"
// @Param       purpose formData string true "Purpose"
// @Success     200 {object} models.MediaUploadResponse
// @Failure     400 {object} map[string]any "VALIDATION, MEDIA_TYPE_NOT_ALLOWED, IMAGE_TOO_SMALL"
// @Failure     413 {object} map[string]any "MEDIA_TOO_LARGE"
// @Failure     503 {object} map[string]any "STORAGE_UNAVAILABLE"
// @Router      /media/upload [post]
func (h *MediaAPIHandler) MediaUpload(c *fiber.Ctx) error {
	request := &models.MediaUploadRequest{Purpose: c.FormValue("purpose")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		if fh, err := c.FormFile("file"); err == nil {
			if fh.Size > h.maxBytes {
				return nil, cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
			}
			f, err := fh.Open()
			if err != nil {
				return nil, cerrors.Wrap(err, "open upload")
			}
			defer f.Close()
			if request.Data, err = io.ReadAll(io.LimitReader(f, h.maxBytes+1)); err != nil {
				return nil, cerrors.Wrap(err, "read upload")
			}
			request.FileName = fh.Filename
		}
		return h.mediaUseCase.Upload(ctx, request)
	})
}
