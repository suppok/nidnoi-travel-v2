package sitebuild

// Values of site_build.trigger and site_build.outcome. MUST match the CHECK constraints in
// migrations/*_create_site_build and frontend/backoffice/src/constants/siteBuilds.ts.
const (
	TriggerManual = "manual" // "rebuild now" button (POST /site/rebuild)
	TriggerAuto   = "auto"   // debounced content edits
	TriggerDaily  = "daily"  // scheduled daily rebuild only

	OutcomeTriggered = "triggered" // Cloudflare accepted the deploy hook
	OutcomeFailed    = "failed"    // every attempt failed
	OutcomeSkipped   = "skipped"   // CLOUDFLARE_DEPLOY_HOOK_URL not set
)

var Triggers = []string{TriggerManual, TriggerAuto, TriggerDaily}

// SystemUserID is create_by for builds no user started (auto, daily), like seeded rows.
const SystemUserID = "00000000-0000-0000-0000-000000000000"
