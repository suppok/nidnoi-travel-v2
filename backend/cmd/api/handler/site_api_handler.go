package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/api/middleware"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/site"
	"backend/pkg/modules/site/models"
)

type SiteAPIHandler struct {
	router      fiber.Router
	apiHandler  *api.ApiHandler
	siteUseCase site.SiteUseCase
}

func NewSiteAPIHandler(router fiber.Router, apiHandler *api.ApiHandler, siteUseCase site.SiteUseCase) *SiteAPIHandler {
	return &SiteAPIHandler{router: router, apiHandler: apiHandler, siteUseCase: siteUseCase}
}

func (h *SiteAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
	h.router.Post("/site/rebuild", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.SiteRebuild), h.SiteRebuild)
	h.router.Post("/site-builds/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.SiteBuildList), h.SiteBuildList)
}

// SiteBuildList godoc
// @Summary     List customer site builds triggered by the backend
// @Description Manual ("rebuild now"), auto (debounced content edits) and daily builds, newest first.
// @Description Builds from git pushes are not included. Outcome is whether Cloudflare accepted the
// @Description trigger, not whether the build itself succeeded.
// @Tags        site
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body models.SiteBuildListRequest true "Filter"
// @Success     200 {object} models.SiteBuildListResponse
// @Failure     400 {object} map[string]any "VALIDATION"
// @Router      /site-builds/list [post]
func (h *SiteAPIHandler) SiteBuildList(c *fiber.Ctx) error {
	request := &models.SiteBuildListRequest{}
	return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.siteUseCase.BuildList(ctx, request)
	})
}

// SiteRebuild godoc
// @Summary     Rebuild the customer site now
// @Description Fires the Cloudflare deploy hook immediately, bypassing the debounce. The build takes
// @Description a few minutes; a failed build leaves the previous deployment live.
// @Tags        site
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} models.SiteRebuildResponse
// @Failure     502 {object} map[string]any "SITE_REBUILD_FAILED"
// @Failure     503 {object} map[string]any "SITE_REBUILD_UNAVAILABLE"
// @Router      /site/rebuild [post]
func (h *SiteAPIHandler) SiteRebuild(c *fiber.Ctx) error {
	request := &models.SiteRebuildRequest{}
	return h.apiHandler.Do(c, nil, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error) {
		request.UserRequestInfo = requestInfo.UserRequestInfo
		return h.siteUseCase.Rebuild(ctx, request)
	})
}
