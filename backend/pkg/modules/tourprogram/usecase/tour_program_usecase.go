package usecase

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogram"
	"backend/pkg/modules/tourprogram/models"
	"backend/pkg/registry"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxCountries    = 20
	maxCities       = 50
	maxImages       = 30
	maxDays         = 60
)

// Storage key prefixes per media field; they must match the media module's purposes.
const (
	bannerPrefix   = "tours/banners/"
	cardPrefix     = "tours/cards/"
	galleryPrefix  = "tours/gallery/"
	documentPrefix = "tours/documents/"
)

type tourProgramUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewTourProgramUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) tourprogram.TourProgramUseCase {
	return &tourProgramUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// normalize trims text and drops empty media values so "" and null mean the same.
func normalize(f *models.TourProgramFields) {
	f.Name = strings.TrimSpace(f.Name)
	f.WholesaleTourCode = strings.TrimSpace(f.WholesaleTourCode)
	if v := strings.TrimSpace(f.AirlineID.String); f.AirlineID.Valid && v != "" {
		f.AirlineID = null.StringFrom(v)
	} else {
		f.AirlineID = null.String{}
	}
	f.Highlight = strings.TrimSpace(f.Highlight)
	f.Description = strings.TrimSpace(f.Description)
	f.InternalNote = strings.TrimSpace(f.InternalNote)
	for _, s := range []*null.String{&f.BannerImage, &f.CardImage, &f.PDFFile, &f.DOCXFile} {
		if v := strings.TrimSpace(s.String); s.Valid && v != "" {
			*s = null.StringFrom(v)
		} else {
			*s = null.String{}
		}
	}
	if f.State == 0 {
		f.State = tourprogramstate.Draft
	}
	if f.CountryIDs == nil {
		f.CountryIDs = []string{}
	}
	if f.CityIDs == nil {
		f.CityIDs = []string{}
	}
	if f.Images == nil {
		f.Images = []string{}
	}
}

func knownStates() []any {
	known := make([]any, len(tourprogramstate.All))
	for i, s := range tourprogramstate.All {
		known[i] = s
	}
	return known
}

func optionalUUID(value any) error {
	v, _ := value.(null.String)
	if !v.Valid {
		return nil
	}
	return validation.Validate(v.String, validation.UUID)
}

func uniqueIDs(value any) error {
	ids, _ := value.([]string)
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return errors.New("must not contain duplicates")
		}
		seen[id] = true
	}
	return nil
}

// mediaURL accepts only URLs this backend uploaded: <R2_PUBLIC_URL>/<prefix>...<ext>.
func (u *tourProgramUseCase) mediaURL(prefix, ext string) validation.Rule {
	return validation.By(func(value any) error {
		var s string
		switch v := value.(type) {
		case null.String:
			if !v.Valid {
				return nil
			}
			s = v.String
		case string:
			s = v
		}
		base := u.shareRegistry.ClientRegistry.ObjectStorage.PublicBaseURL()
		if base == "" {
			return errors.New("file storage is not configured")
		}
		rest, ok := strings.CutPrefix(s, base+"/"+prefix)
		if !ok || rest == "" || strings.ContainsAny(rest, "?#") || strings.Contains(rest, "..") {
			return errors.New("must be a file uploaded through /media/upload")
		}
		if ext != "" && !strings.HasSuffix(strings.ToLower(rest), ext) {
			return errors.New("must be a " + strings.TrimPrefix(ext, ".") + " file")
		}
		return nil
	})
}

// fieldsValidate holds the field rules shared by create and update.
func (u *tourProgramUseCase) fieldsValidate(f *models.TourProgramFields) error {
	return validation.ValidateStruct(f,
		validation.Field(&f.RefID, validation.By(func(value any) error {
			if id, ok := value.(null.Int); ok && id.Valid && id.Int64 < 1 {
				return errors.New("must be a positive number")
			}
			return nil
		})),
		validation.Field(&f.Name, validation.Required, validation.RuneLength(1, 200)),
		validation.Field(&f.WholesaleID, validation.Required, validation.UUID),
		validation.Field(&f.WholesaleTourCode, validation.RuneLength(0, 50)),
		validation.Field(&f.State, validation.Required, validation.In(knownStates()...)),
		validation.Field(&f.Days, validation.Required, validation.Min(1), validation.Max(maxDays)),
		validation.Field(&f.Nights, validation.Min(0), validation.Max(f.Days).Error("must not be more than days")),
		validation.Field(&f.AirlineID, validation.By(optionalUUID)),
		validation.Field(&f.Highlight, validation.RuneLength(0, 500)),
		validation.Field(&f.Description, validation.RuneLength(0, 20000)),
		validation.Field(&f.BannerImage, u.mediaURL(bannerPrefix, "")),
		validation.Field(&f.CardImage, u.mediaURL(cardPrefix, "")),
		validation.Field(&f.PDFFile, u.mediaURL(documentPrefix, ".pdf")),
		validation.Field(&f.DOCXFile, u.mediaURL(documentPrefix, ".docx")),
		validation.Field(&f.InternalNote, validation.RuneLength(0, 5000)),
		validation.Field(&f.CountryIDs, validation.Required.Error("select at least one country"),
			validation.Length(1, maxCountries), validation.Each(validation.UUID), validation.By(uniqueIDs)),
		validation.Field(&f.CityIDs, validation.Length(0, maxCities), validation.Each(validation.UUID), validation.By(uniqueIDs)),
		validation.Field(&f.Images, validation.Length(0, maxImages), validation.Each(u.mediaURL(galleryPrefix, "")), validation.By(uniqueIDs)),
	)
}

// checkReferences verifies the wholesale, airline, countries and cities inside the transaction. Newly
// selected references must be active; ones the tour already had may stay even if deleted since.
// previous is nil on create.
func (u *tourProgramUseCase) checkReferences(ctx context.Context, appDB query.AppIDB, f *models.TourProgramFields, previous *domain.TourProgram) error {
	ws, err := u.repositoryRegistry.WholesaleRepository.WithTx(appDB).FindByID(ctx, f.WholesaleID)
	if err != nil {
		return err
	}
	if ws == nil || (!ws.Status && (previous == nil || previous.WholesaleID != f.WholesaleID)) {
		return cerrors.FieldError("wholesaleId", "wholesale not found or deleted")
	}

	if f.AirlineID.Valid {
		al, err := u.repositoryRegistry.AirlineRepository.WithTx(appDB).FindByID(ctx, f.AirlineID.String)
		if err != nil {
			return err
		}
		kept := previous != nil && previous.AirlineID.Valid && previous.AirlineID.String == f.AirlineID.String
		if al == nil || (!al.Status && !kept) {
			return cerrors.FieldError("airlineId", "airline not found or deleted")
		}
	}

	hadCountry, hadCity := map[string]bool{}, map[string]bool{}
	if previous != nil {
		for _, l := range previous.CountryLinks {
			hadCountry[l.CountryID] = true
		}
		for _, l := range previous.CityLinks {
			hadCity[l.CityID] = true
		}
	}

	countries, err := u.repositoryRegistry.CountryRepository.WithTx(appDB).FindByIDs(ctx, f.CountryIDs)
	if err != nil {
		return err
	}
	found := make(map[string]bool, len(countries))
	for _, c := range countries {
		if c.Status || hadCountry[c.ID] {
			found[c.ID] = true
		}
	}
	for _, id := range f.CountryIDs {
		if !found[id] {
			return cerrors.FieldError("countryIds", "a country was not found or is deleted")
		}
	}

	cities, err := u.repositoryRegistry.CityRepository.WithTx(appDB).FindByIDs(ctx, f.CityIDs)
	if err != nil {
		return err
	}
	cityOK := make(map[string]bool, len(cities))
	for _, c := range cities {
		if !slices.Contains(f.CountryIDs, c.CountryID) {
			return cerrors.FieldError("cityIds", "every city must belong to one of the selected countries")
		}
		if c.Status || hadCity[c.ID] {
			cityOK[c.ID] = true
		}
	}
	for _, id := range f.CityIDs {
		if !cityOK[id] {
			return cerrors.FieldError("cityIds", "a city was not found or is deleted")
		}
	}
	return nil
}

// replaceChildren swaps the tour's country/city links and gallery for the request's lists.
func (u *tourProgramUseCase) replaceChildren(ctx context.Context, appDB query.AppIDB, tourID, actorID string, f *models.TourProgramFields) error {
	countryLinks := make([]domain.TourProgramCountry, len(f.CountryIDs))
	for i, id := range f.CountryIDs {
		countryLinks[i] = domain.TourProgramCountry{TourProgramID: tourID, CountryID: id, SortOrder: i}
	}
	if err := u.repositoryRegistry.TourProgramCountryRepository.WithTx(appDB).ReplaceForTour(ctx, tourID, countryLinks); err != nil {
		return err
	}

	cityLinks := make([]domain.TourProgramCity, len(f.CityIDs))
	for i, id := range f.CityIDs {
		cityLinks[i] = domain.TourProgramCity{TourProgramID: tourID, CityID: id, SortOrder: i}
	}
	if err := u.repositoryRegistry.TourProgramCityRepository.WithTx(appDB).ReplaceForTour(ctx, tourID, cityLinks); err != nil {
		return err
	}

	now := time.Now()
	images := make([]domain.TourProgramImage, len(f.Images))
	for i, url := range f.Images {
		images[i] = domain.TourProgramImage{
			ID:            utils.NewID(),
			TourProgramID: tourID,
			URL:           url,
			SortOrder:     i,
			CreateDate:    now,
			UpdateDate:    now,
			CreateBy:      actorID,
			UpdateBy:      actorID,
		}
	}
	return u.repositoryRegistry.TourProgramImageRepository.WithTx(appDB).ReplaceForTour(ctx, tourID, images)
}

// applyFields copies the editable fields onto the row (everything except id, code and audit).
func applyFields(t *domain.TourProgram, f *models.TourProgramFields) {
	t.RefID = f.RefID
	t.Name = f.Name
	t.WholesaleID = f.WholesaleID
	t.WholesaleTourCode = f.WholesaleTourCode
	t.State = f.State
	t.Days = f.Days
	t.Nights = f.Nights
	t.AirlineID = f.AirlineID
	t.Highlight = f.Highlight
	t.Description = f.Description
	t.BannerImage = f.BannerImage
	t.CardImage = f.CardImage
	t.PDFFile = f.PDFFile
	t.DOCXFile = f.DOCXFile
	t.InternalNote = f.InternalNote
}

var editableColumns = []string{
	"ref_id", "name", "wholesale_id", "wholesale_tour_code", "state", "days", "nights", "airline_id",
	"highlight", "description", "banner_image", "card_image", "pdf_file", "docx_file", "internal_note",
	"update_date", "update_by",
}

// uniqueConflict maps a unique-index violation to REF_ID_TAKEN; other errors pass through.
func uniqueConflict(err error) error {
	if query.UniqueViolationConstraint(err) == "tour_program_ref_id_key" {
		return cerrors.Conflict(errorcode.RefIDTaken)
	}
	return err
}

// load reads the tour with every relation the item needs, or fails with NOT_FOUND.
func (u *tourProgramUseCase) load(ctx context.Context, appDB query.AppIDB, id string) (*domain.TourProgram, error) {
	found, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return found, nil
}

// stripRelations clears relation fields before an Update so Bun writes only columns.
func stripRelations(t *domain.TourProgram) {
	t.Wholesale = nil
	t.Airline = nil
	t.CountryLinks = nil
	t.CityLinks = nil
	t.Images = nil
}
