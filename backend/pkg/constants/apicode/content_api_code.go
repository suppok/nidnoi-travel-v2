package apicode

// Declared ahead of their endpoints so /auth/me already reports them to the backoffice,
// which mirrors them in src/constants/permissions.ts.

var (
	TourProgramCreate = ApiCode("tour-program:create", "Create a tour program")
	TourProgramUpdate = ApiCode("tour-program:update", "Update a tour program")
	TourProgramList   = ApiCode("tour-program:list", "List tour programs")
	TourProgramGet    = ApiCode("tour-program:get", "Get a tour program")
	TourProgramStatus = ApiCode("tour-program:status", "Delete or restore a tour program")
	PeriodCreate      = ApiCode("tour-program-period:create", "Add a period to a tour program")
	PeriodUpdate      = ApiCode("tour-program-period:update", "Update periods, prices and seats")
	PeriodList        = ApiCode("tour-program-period:list", "List the periods of a tour program")
	PeriodStatus      = ApiCode("tour-program-period:status", "Delete or restore a period")
	CountryCreate     = ApiCode("country:create", "Create a country")
	CountryUpdate     = ApiCode("country:update", "Update a country")
	CountryStatus     = ApiCode("country:status", "Delete or restore a country")
	CountryList       = ApiCode("country:list", "List countries")
	CountryGet        = ApiCode("country:get", "Get a country")
	CityCreate        = ApiCode("city:create", "Create a city")
	CityUpdate        = ApiCode("city:update", "Update a city")
	CityStatus        = ApiCode("city:status", "Delete or restore a city")
	CityList          = ApiCode("city:list", "List cities")
	CityGet           = ApiCode("city:get", "Get a city")
	MediaUpload       = ApiCode("media:upload", "Upload media")
	SiteRebuild       = ApiCode("site:rebuild", "Trigger a customer site rebuild")
	SiteBuildList     = ApiCode("site-build:list", "View the customer site build history")
)
