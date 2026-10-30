package apicode

var (
	WholesaleCreate = ApiCode("wholesale:create", "Create a wholesale")
	WholesaleUpdate = ApiCode("wholesale:update", "Update a wholesale and its contacts")
	WholesaleList   = ApiCode("wholesale:list", "List wholesales")
	WholesaleGet    = ApiCode("wholesale:get", "Get a wholesale")
	WholesaleStatus = ApiCode("wholesale:status", "Delete or restore a wholesale")
)
