package apicode

var (
	// AISettingGet / AISettingUpdate: the Claude API key, model and on/off for AI features (admins via "*").
	// POST /ai-settings/test uses AISettingUpdate.
	AISettingGet    = ApiCode("ai-setting:get", "View the AI settings")
	AISettingUpdate = ApiCode("ai-setting:update", "Change the AI settings and test the API key")
)
