package usecase

import (
	"backend/pkg/modules/media"
	"backend/pkg/modules/media/commons"
	"backend/pkg/modules/media/models"
	"backend/pkg/registry"
)

// purposeRule says what a purpose accepts and where it is stored.
type purposeRule struct {
	mimes     []string
	minWidth  int // images only
	minHeight int
	keyPrefix string
}

var imageMimes = []string{commons.MimeJPEG, commons.MimePNG, commons.MimeWebP}

var purposeRules = map[string]purposeRule{
	models.PurposeTourBanner:  {mimes: imageMimes, minWidth: 1200, minHeight: 630, keyPrefix: "tours/banners/"},
	models.PurposeTourCard:    {mimes: imageMimes, minWidth: 800, minHeight: 600, keyPrefix: "tours/cards/"},
	models.PurposeTourGallery: {mimes: imageMimes, minWidth: 800, minHeight: 600, keyPrefix: "tours/gallery/"},
	models.PurposeTourPDF:     {mimes: []string{commons.MimePDF}, keyPrefix: "tours/documents/"},
	models.PurposeTourDOCX:    {mimes: []string{commons.MimeDOCX}, keyPrefix: "tours/documents/"},
	models.PurposeCountryHero: {mimes: imageMimes, minWidth: 1600, minHeight: 900, keyPrefix: "countries/heroes/"},
	models.PurposeCityHero:    {mimes: imageMimes, minWidth: 1600, minHeight: 900, keyPrefix: "cities/heroes/"},
	// Logos are small and often wide (e.g. 400×100); airline usecase checks this prefix.
	models.PurposeAirlineLogo: {mimes: imageMimes, minWidth: 120, minHeight: 40, keyPrefix: "airlines/logos/"},
}

func (r purposeRule) isImage() bool { return r.minWidth > 0 }

type mediaUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewMediaUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) media.MediaUseCase {
	return &mediaUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}
