package usecase

import (
	"slices"

	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

// seeAllRoles see every LINE booking; any other staff user sees only the bookings they own.
var seeAllRoles = []string{role.SuperAdmin, role.Admin, role.Finance}

// ownerScope returns the user id the caller is limited to, or "" when they see every booking.
func ownerScope(info *coremodels.UserRequestInfo) string {
	if info == nil {
		return ""
	}
	for _, r := range info.Roles {
		if slices.Contains(seeAllRoles, r) {
			return ""
		}
	}
	return info.UserID
}

// visibleTo reports whether the caller may see b. Hidden bookings are reported as NOT_FOUND so
// their existence is not revealed.
func visibleTo(b *domain.LineBooking, info *coremodels.UserRequestInfo) bool {
	owner := ownerScope(info)
	return owner == "" || (b.SalesUserID.Valid && b.SalesUserID.String == owner)
}
