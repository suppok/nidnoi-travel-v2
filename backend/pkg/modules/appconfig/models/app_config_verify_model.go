package models

import coremodels "backend/pkg/core/api/models"

type AppConfigVerifyRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// AppConfigVerifyIssue is one config that needs fixing.
type AppConfigVerifyIssue struct {
	Code string `json:"code"`
	// NOT_FOUND: a required code has no active row. INVALID: the stored value doesn't decode
	// into the code's type or fails its rules (message says why).
	Type    string `json:"type" enums:"NOT_FOUND,INVALID"`
	Message string `json:"message"`
}

// AppConfigVerifyResponse is the list of issues, ordered by code; empty when everything is fine.
type AppConfigVerifyResponse []AppConfigVerifyIssue
