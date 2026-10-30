package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// WholesaleContactInput is one contact person in a create/update body, in display order.
type WholesaleContactInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type WholesaleCreateRequest struct {
	RefID    null.Int                `json:"refId" swaggertype:"integer"` // optional legacy id
	Code     string                  `json:"code"`                        // stored uppercase, unique
	Name     string                  `json:"name"`
	State    int                     `json:"state"` // default 10 (on request)
	Contacts []WholesaleContactInput `json:"contacts"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type WholesaleCreateResponse struct {
	WholesaleItem
}
