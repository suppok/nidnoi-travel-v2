package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// WholesaleUpdateRequest replaces every field, including the whole contact list.
type WholesaleUpdateRequest struct {
	ID       string                  `json:"id"`
	RefID    null.Int                `json:"refId" swaggertype:"integer"`
	Code     string                  `json:"code"`
	Name     string                  `json:"name"`
	State    int                     `json:"state"`
	Contacts []WholesaleContactInput `json:"contacts"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type WholesaleUpdateResponse struct {
	WholesaleItem
}
