package apicode

var (
	// AppConfigVerify is also limited to the super admin in the use case (admins hold "*").
	AppConfigVerify      = ApiCode("app-config:verify", "Verify that every app config is present and valid")
	AppConfigAppSettings = ApiCode("app-config:app-settings", "View the general app settings")
	// AppConfigEmailContactCard: view and change the contact card image in email footers (admins via "*").
	AppConfigEmailContactCard = ApiCode("app-config:email-contact-card", "Manage the contact card image at the bottom of emails")
)
