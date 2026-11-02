package models

import coremodels "backend/pkg/core/api/models"

// --- create ---

type TourProgramPeriodCreateRequest struct {
	TourProgramID string `json:"tourProgramId"`
	TourProgramPeriodFields

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramPeriodCreateResponse struct {
	TourProgramPeriodItem
}

// --- update ---

type TourProgramPeriodUpdateRequest struct {
	ID string `json:"id"`
	TourProgramPeriodFields

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramPeriodUpdateResponse struct {
	TourProgramPeriodItem
}

// --- list ---

type TourProgramPeriodListRequest struct {
	TourProgramID string `json:"tourProgramId"`
	Status        *bool  `json:"status"` // omitted = all

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// TourProgramPeriodListResponse lists every period of one tour by departure date (no paging).
type TourProgramPeriodListResponse struct {
	Items []TourProgramPeriodItem `json:"items"`
}

// --- status ---

type TourProgramPeriodUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// TourProgramPeriodUpdateStatusResponse is always nil (data: null).
type TourProgramPeriodUpdateStatusResponse struct{}
