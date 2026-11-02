package models

import coremodels "backend/pkg/core/api/models"

// Upload purposes (form field "purpose").
const (
	PurposeTourBanner  = "tour-banner"
	PurposeTourCard    = "tour-card"
	PurposeTourGallery = "tour-gallery"
	PurposeTourPDF     = "tour-pdf"
	PurposeTourDOCX    = "tour-docx"
	PurposeCountryHero = "country-hero"
	PurposeCityHero    = "city-hero"
)

// MediaUploadRequest is built by the handler from multipart/form-data (fields "file", "purpose").
type MediaUploadRequest struct {
	Purpose  string `json:"-"`
	FileName string `json:"-"` // client file name, used only for the download name of documents
	Data     []byte `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type MediaUploadResponse struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int    `json:"size"`
	Mime     string `json:"mime"`
	FileName string `json:"fileName"` // download name for documents; "" for images
	Width    int    `json:"width"`    // 0 for documents
	Height   int    `json:"height"`   // 0 for documents
}
