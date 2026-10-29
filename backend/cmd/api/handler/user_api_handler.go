package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/user"
	"backend/pkg/modules/user/models"
)

type UserAPIHandler struct {
	router      fiber.Router
	apiHandler  *api.ApiHandler
	userUseCase user.UserUseCase
}

func NewUserAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, userUseCase user.UserUseCase) *UserAPIHandler {
	return &UserAPIHandler{router: router, apiHandler: apiHandler, userUseCase: userUseCase}
}

func (h *UserAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	r := h.router

	// --- Staff (backoffice) — fixed paths before "/:id". ---
	r.Post("/users", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserCreate), h.UserCreate)
	r.Put("/users", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserUpdate), h.UserUpdate)
	r.Post("/users/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserList), h.UserList)
	r.Put("/users/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserStatus), h.UserUpdateStatus)
	r.Put("/users/password", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserPasswordReset), h.UserResetPassword)
	r.Get("/users/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.UserGet), h.UserGet)
}

// UserList godoc
// @Summary  List staff users
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.UserListRequest true "Filter"
// @Success  200 {object} models.UserListResponse
// @Router   /users/list [post]
func (h *UserAPIHandler) UserList(c *fiber.Ctx) error {
	request := &models.UserListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.List(ctx, request)
	})
}

// UserGet godoc
// @Summary  Get a staff user
// @Tags     users
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "User ID"
// @Success  200 {object} models.UserGetResponse
// @Failure  404 {object} map[string]any "NOT_FOUND"
// @Router   /users/{id} [get]
func (h *UserAPIHandler) UserGet(c *fiber.Ctx) error {
	request := &models.UserGetRequest{ID: c.Params("id")}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.Get(ctx, request)
	})
}

// UserCreate godoc
// @Summary     Create a staff user
// @Description The user must replace the temporary password at first login.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.UserCreateRequest true "User"
// @Success     200 {object} models.UserCreateResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Failure     409 {object} map[string]any "EMAIL_TAKEN"
// @Failure     403 {object} map[string]any "SUPER_ADMIN_REQUIRED (admin role)"
// @Router      /users [post]
func (h *UserAPIHandler) UserCreate(c *fiber.Ctx) error {
	request := &models.UserCreateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.Create(ctx, request)
	})
}

// UserUpdate godoc
// @Summary  Update a staff user's email, name and roles
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body body models.UserUpdateRequest true "User"
// @Success  200 {object} models.UserUpdateResponse
// @Failure  409 {object} map[string]any "EMAIL_TAKEN, CANNOT_MODIFY_SELF, SUPER_ADMIN_PROTECTED"
// @Failure  403 {object} map[string]any "SUPER_ADMIN_REQUIRED"
// @Router   /users [put]
func (h *UserAPIHandler) UserUpdate(c *fiber.Ctx) error {
	request := &models.UserUpdateRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.Update(ctx, request)
	})
}

// UserUpdateStatus godoc
// @Summary     Deactivate or restore a staff user
// @Description Deactivation signs the user out everywhere.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.UserUpdateStatusRequest true "Status"
// @Success     200 {object} map[string]any "data: null"
// @Failure     409 {object} map[string]any "CANNOT_MODIFY_SELF, SUPER_ADMIN_PROTECTED"
// @Failure     403 {object} map[string]any "SUPER_ADMIN_REQUIRED"
// @Router      /users/status [put]
func (h *UserAPIHandler) UserUpdateStatus(c *fiber.Ctx) error {
	request := &models.UserUpdateStatusRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.UpdateStatus(ctx, request)
	})
}

// UserResetPassword godoc
// @Summary     Set a temporary password for a staff user
// @Description Forces a password change at next login and signs the user out everywhere.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.UserResetPasswordRequest true "Temporary password"
// @Success     200 {object} map[string]any "data: null"
// @Failure     409 {object} map[string]any "CANNOT_MODIFY_SELF, SUPER_ADMIN_PROTECTED"
// @Failure     403 {object} map[string]any "SUPER_ADMIN_REQUIRED"
// @Router      /users/password [put]
func (h *UserAPIHandler) UserResetPassword(c *fiber.Ctx) error {
	request := &models.UserResetPasswordRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.userUseCase.ResetPassword(ctx, request)
	})
}
