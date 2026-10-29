package role

// Fixed roles. MUST match config/permissions/policy.csv, the CHECK constraint on app_user.roles
// and frontend/backoffice/src/constants/roles.ts.
const (
	// SuperAdmin is the first user, created only by `migrate create-super-admin`. At most one
	// exists (unique index). It can't be granted, removed or changed through the API by anyone else.
	SuperAdmin = "super_admin"
	Admin      = "admin"
	Sales      = "sales"
	Finance    = "finance"
)

var All = []string{SuperAdmin, Admin, Sales, Finance}

// Assignable are the roles staff can grant through the API.
var Assignable = []string{Admin, Sales, Finance}

func Valid(r string) bool {
	for _, v := range All {
		if v == r {
			return true
		}
	}
	return false
}
