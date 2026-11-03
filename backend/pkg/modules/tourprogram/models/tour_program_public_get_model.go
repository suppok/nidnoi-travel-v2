package models

type TourProgramPublicGetRequest struct {
	Code string // from the path
}

type TourProgramPublicGetResponse struct {
	TourProgramPublicDetail
}
