package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/errorcode"
	coremodels "backend/pkg/core/api/models"
	cerrors "backend/pkg/core/errors"
)

const userRequestInfoKey = "userRequestInfo"

// HandlerFunc is the body of a handler: call one use case method and return its response.
type HandlerFunc func(ctx context.Context, requestInfo *coremodels.RequestInfo) (any, error)

// ApiHandler binds the request, builds RequestInfo and writes the envelope.
type ApiHandler struct {
	response       *ApiResponseHandler
	clientIPHeader string
}

func NewApiHandler(response *ApiResponseHandler, clientIPHeader string) *ApiHandler {
	return &ApiHandler{response: response, clientIPHeader: clientIPHeader}
}

// Do parses a JSON body into request (when request is non-nil and the body is non-empty),
// then runs fn and writes its result.
func (h *ApiHandler) Do(c *fiber.Ctx, request any, fn HandlerFunc) error {
	if request != nil && len(c.Body()) > 0 {
		if err := c.BodyParser(request); err != nil {
			return h.response.Error(c, &cerrors.AppError{
				Status:  http.StatusBadRequest,
				Code:    errorcode.Validation.Code,
				Message: "Malformed JSON body",
			})
		}
	}

	info := &coremodels.RequestInfo{
		UserRequestInfo: UserRequestInfoFrom(c),
		ClientIP:        ClientIP(c, h.clientIPHeader),
		UserAgent:       string(c.Request().Header.UserAgent()),
	}

	data, err := fn(c.UserContext(), info)
	if err != nil {
		return h.response.Error(c, err)
	}
	return h.response.Success(c, data)
}

// SetUserRequestInfo is called by the auth middleware.
func SetUserRequestInfo(c *fiber.Ctx, info *coremodels.UserRequestInfo) {
	c.Locals(userRequestInfoKey, info)
}

func UserRequestInfoFrom(c *fiber.Ctx) *coremodels.UserRequestInfo {
	info, _ := c.Locals(userRequestInfoKey).(*coremodels.UserRequestInfo)
	return info
}

// ClientIP prefers the proxy header (Cloudflare's CF-Connecting-IP) and falls back to the socket address.
func ClientIP(c *fiber.Ctx, header string) string {
	if header != "" {
		if ip := c.Get(header); ip != "" {
			return ip
		}
	}
	return c.IP()
}
