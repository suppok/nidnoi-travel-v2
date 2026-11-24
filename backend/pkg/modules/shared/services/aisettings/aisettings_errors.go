package aisettings

import (
	"errors"
	"net/http"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/shared/client/anthropic"
)

// NotConfigured is AI_NOT_CONFIGURED (503).
func NotConfigured(reason string) *cerrors.AppError {
	e := cerrors.New(http.StatusServiceUnavailable, errorcode.AINotConfigured)
	if reason != "" {
		e.Message += ": " + reason
	}
	return e
}

// MapClientError turns an anthropic client error into the API error staff see. The cause is kept
// for the log; it never contains the key or the document.
func MapClientError(err error) *cerrors.AppError {
	var e *cerrors.AppError
	var req *anthropic.RequestError
	switch {
	case errors.Is(err, anthropic.ErrInvalidAPIKey):
		e = cerrors.New(http.StatusBadRequest, errorcode.AIInvalidAPIKey)
	case errors.Is(err, anthropic.ErrModelNotFound):
		e = cerrors.New(http.StatusBadRequest, errorcode.AIModelUnavailable)
	case errors.Is(err, anthropic.ErrBusy):
		e = cerrors.New(http.StatusServiceUnavailable, errorcode.AIBusy)
	case errors.Is(err, anthropic.ErrTimeout):
		e = cerrors.New(http.StatusGatewayTimeout, errorcode.AIRequestFailed)
		e.Message += ": the AI took too long"
	case errors.Is(err, anthropic.ErrRefused):
		e = cerrors.New(http.StatusBadGateway, errorcode.AIRequestFailed)
		e.Message += ": the AI declined to read this document"
	case errors.Is(err, anthropic.ErrTruncated):
		e = cerrors.New(http.StatusBadGateway, errorcode.AIRequestFailed)
		e.Message += ": the answer was too long"
	case errors.As(err, &req):
		e = cerrors.New(http.StatusBadGateway, errorcode.AIRequestFailed)
		if req.Message != "" {
			e.Message += ": " + req.Message
		}
	default:
		e = cerrors.New(http.StatusBadGateway, errorcode.AIRequestFailed)
	}
	e.Err = err
	return e
}
