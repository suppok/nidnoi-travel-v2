package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/config"
	"backend/pkg/modules/auth"
	"backend/pkg/modules/auth/models"
)

// The refresh cookie is only sent to /api/v1/auth/* (refresh, logout, password).
const refreshCookiePath = "/api/v1/auth"

type AuthAPIHandler struct {
	router      fiber.Router
	apiHandler  *api.ApiHandler
	config      *config.Config
	authUseCase auth.AuthUseCase
}

func NewAuthAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, cfg *config.Config, authUseCase auth.AuthUseCase) *AuthAPIHandler {
	return &AuthAPIHandler{router: router, apiHandler: apiHandler, config: cfg, authUseCase: authUseCase}
}

func (h *AuthAPIHandler) Init(jwtAuth middleware.Auth) {
	r := h.router

	r.Post("/auth/login", middleware.RateLimit(h.config.RateLimit.Login, h.config.Server.ClientIPHeader), h.AuthLogin)
	r.Post("/auth/refresh", h.AuthRefresh)
	r.Post("/auth/logout", h.AuthLogout)
	r.Get("/auth/me", jwtAuth.AuthAllowPasswordChange, h.AuthMe)
	r.Put("/auth/password", jwtAuth.AuthAllowPasswordChange, h.AuthChangePassword)
}

// AuthLogin godoc
// @Summary     Log in (backoffice)
// @Description Returns a short-lived access token and sets the HttpOnly refresh cookie. Rate-limited.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body models.AuthLoginRequest true "Credentials"
// @Success     200 {object} models.AuthLoginResponse
// @Failure     401 {object} map[string]any "INVALID_CREDENTIALS"
// @Failure     429 {object} map[string]any "RATE_LIMITED"
// @Router      /auth/login [post]
func (h *AuthAPIHandler) AuthLogin(c *fiber.Ctx) error {
	request := &models.AuthLoginRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.ClientIP = requestInfo.ClientIP
		request.UserAgent = requestInfo.UserAgent
		response, err := h.authUseCase.Login(ctx, request)
		if err != nil {
			return nil, err
		}
		h.setRefreshCookie(c, response.RefreshToken, response.RefreshExpiresAt, response.Persistent)
		return response, nil
	})
}

// AuthRefresh godoc
// @Summary     Refresh the access token
// @Description Uses and rotates the refresh cookie. 401 clears the cookie.
// @Tags        auth
// @Produce     json
// @Success     200 {object} models.AuthRefreshResponse
// @Failure     401 {object} map[string]any "UNAUTHENTICATED"
// @Router      /auth/refresh [post]
func (h *AuthAPIHandler) AuthRefresh(c *fiber.Ctx) error {
	request := &models.AuthRefreshRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.RefreshToken = c.Cookies(h.config.Auth.RefreshCookieName)
		request.ClientIP = requestInfo.ClientIP
		request.UserAgent = requestInfo.UserAgent
		response, err := h.authUseCase.Refresh(ctx, request)
		if err != nil {
			h.clearRefreshCookie(c)
			return nil, err
		}
		h.setRefreshCookie(c, response.RefreshToken, response.RefreshExpiresAt, response.Persistent)
		return response, nil
	})
}

// AuthLogout godoc
// @Summary     Log out
// @Description Revokes the refresh cookie's token and clears the cookie. No access token needed.
// @Tags        auth
// @Produce     json
// @Success     200 {object} map[string]any "data: null"
// @Router      /auth/logout [post]
func (h *AuthAPIHandler) AuthLogout(c *fiber.Ctx) error {
	request := &models.AuthLogoutRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, _ *coremodels.RequestInfo) (any, error) {
		request.RefreshToken = c.Cookies(h.config.Auth.RefreshCookieName)
		h.clearRefreshCookie(c)
		return h.authUseCase.Logout(ctx, request)
	})
}

// AuthMe godoc
// @Summary     Current user
// @Description Roles and the union of their permission codes. Allowed while a password change is pending.
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.AuthMeResponse
// @Failure     401 {object} map[string]any "UNAUTHENTICATED"
// @Router      /auth/me [get]
func (h *AuthAPIHandler) AuthMe(c *fiber.Ctx) error {
	request := &models.AuthMeRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.authUseCase.Me(ctx, request)
	})
}

// AuthChangePassword godoc
// @Summary     Change own password
// @Description Clears mustChangePassword and signs out every other session.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.AuthChangePasswordRequest true "Passwords"
// @Success     200 {object} map[string]any "data: null"
// @Failure     400 {object} map[string]any "VALIDATION, INVALID_PASSWORD"
// @Router      /auth/password [put]
func (h *AuthAPIHandler) AuthChangePassword(c *fiber.Ctx) error {
	request := &models.AuthChangePasswordRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		request.RefreshToken = c.Cookies(h.config.Auth.RefreshCookieName)
		return h.authUseCase.ChangePassword(ctx, request)
	})
}

// setRefreshCookie: persistent ("remember me") cookies carry Expires; others are session cookies.
func (h *AuthAPIHandler) setRefreshCookie(c *fiber.Ctx, token string, expiresAt time.Time, persistent bool) {
	cookie := &fiber.Cookie{
		Name:        h.config.Auth.RefreshCookieName,
		Value:       token,
		Path:        refreshCookiePath,
		HTTPOnly:    true,
		Secure:      !h.config.IsLocal(),
		SameSite:    fiber.CookieSameSiteLaxMode,
		SessionOnly: !persistent,
	}
	if persistent {
		cookie.Expires = expiresAt
	}
	c.Cookie(cookie)
}

func (h *AuthAPIHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     h.config.Auth.RefreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		HTTPOnly: true,
		Secure:   !h.config.IsLocal(),
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
