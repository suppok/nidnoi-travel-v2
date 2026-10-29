package api

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *errorBody `json:"error,omitempty"`
}

// successEnvelope always includes "data", even when it is null.
type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

// ApiResponseHandler writes the cross-project response envelope.
type ApiResponseHandler struct {
	logger *zap.Logger
}

func NewApiResponseHandler(logger *zap.Logger) *ApiResponseHandler {
	return &ApiResponseHandler{logger: logger}
}

func (h *ApiResponseHandler) Success(c *fiber.Ctx, data any) error {
	return c.Status(http.StatusOK).JSON(successEnvelope{Success: true, Data: data})
}

// Error maps *cerrors.AppError to its status and code. Anything else is logged and becomes INTERNAL.
func (h *ApiResponseHandler) Error(c *fiber.Ctx, err error) error {
	var appErr *cerrors.AppError
	if !errors.As(err, &appErr) {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			appErr = fromFiberError(fiberErr)
		} else {
			appErr = cerrors.Wrap(err, "unhandled")
		}
	}

	if appErr.Status >= http.StatusInternalServerError {
		h.logger.Error("request failed",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Error(appErr),
		)
	}

	return c.Status(appErr.Status).JSON(envelope{
		Success: false,
		Error:   &errorBody{Code: appErr.Code, Message: appErr.Message, Fields: appErr.Fields},
	})
}

// FiberErrorHandler is installed as fiber.Config.ErrorHandler so middleware errors, unknown
// routes and oversized bodies also use the envelope.
func (h *ApiResponseHandler) FiberErrorHandler(c *fiber.Ctx, err error) error {
	return h.Error(c, err)
}

func fromFiberError(e *fiber.Error) *cerrors.AppError {
	switch e.Code {
	case http.StatusNotFound:
		return cerrors.NotFound()
	case http.StatusUnauthorized:
		return cerrors.Unauthenticated()
	case http.StatusForbidden:
		return cerrors.PermissionDenied()
	case http.StatusTooManyRequests:
		return cerrors.RateLimited()
	case http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnprocessableEntity, http.StatusMethodNotAllowed:
		return &cerrors.AppError{Status: e.Code, Code: errorcode.Validation.Code, Message: e.Message}
	default:
		return cerrors.Wrap(e, "fiber")
	}
}
