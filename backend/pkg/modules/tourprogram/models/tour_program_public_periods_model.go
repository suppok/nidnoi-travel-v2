package models

type TourProgramPublicPeriodsRequest struct {
	Code string // from the path
}

// TourProgramPublicPeriodsResponse is the live payload of the tour page's period table.
type TourProgramPublicPeriodsResponse struct {
	Items []TourPeriodPublicItem `json:"items"`
}
