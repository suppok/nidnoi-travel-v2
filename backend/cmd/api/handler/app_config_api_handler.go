package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/appconfig"
	"backend/pkg/modules/appconfig/models"
)

type AppConfigAPIHandler struct {
	router           fiber.Router
	apiHandler       *api.ApiHandler
	appConfigUseCase appconfig.AppConfigUseCase
}

func NewAppConfigAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, appConfigUseCase appconfig.AppConfigUseCase) *AppConfigAPIHandler {
	return &AppConfigAPIHandler{router: router, apiHandler: apiHandler, appConfigUseCase: appConfigUseCase}
}

func (h *AppConfigAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router
	r.Post("/app-configs/verify", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AppConfigVerify), h.AppConfigVerify)
	r.Get("/app-configs/app-settings", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.AppConfigAppSettings), h.AppConfigAppSettingsGet)
}

// AppConfigVerify godoc
// @Summary     Verify app configs (super admin only)
// @Description Checks every registered app config code: a required code without an active row is
// @Description NOT_FOUND; a stored value that doesn't decode into the code's type or fails its rules
// @Description is INVALID. An empty list means every config is fine.
// @Tags        app-config
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.AppConfigVerifyResponse
// @Failure     403 {object} map[string]any "SUPER_ADMIN_REQUIRED"
// @Router      /app-configs/verify [post]
func (h *AppConfigAPIHandler) AppConfigVerify(c *fiber.Ctx) error {
	request := &models.AppConfigVerifyRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.appConfigUseCase.Verify(ctx, request)
	})
}

// AppConfigAppSettingsGet godoc
// @Summary     Get the general app settings (APP_SETTINGS)
// @Tags        app-config
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.AppConfigAppSettingsGetResponse
// @Failure     503 {object} map[string]any "APP_CONFIG_NOT_FOUND"
// @Router      /app-configs/app-settings [get]
func (h *AppConfigAPIHandler) AppConfigAppSettingsGet(c *fiber.Ctx) error {
	request := &models.AppConfigAppSettingsGetRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.appConfigUseCase.AppSettingsGet(ctx, request)
	})
}
