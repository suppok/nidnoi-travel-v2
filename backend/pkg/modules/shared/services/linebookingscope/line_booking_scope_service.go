// Package linebookingscope decides which LINE bookings a caller may see. Used by the LINE booking
// module (list, get, actions) and the dashboard (aggregates).
package linebookingscope

import (
	"slices"

	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain/repomodels"
)

// seeAllRoles see every LINE booking. Sales see the bookings they own, finance the bookings sent
// to them on confirm; a user with both roles sees both.
var seeAllRoles = []string{role.SuperAdmin, role.Admin}

// For returns what the caller is limited to, or nil when they see every booking.
func For(info *coremodels.UserRequestInfo) *repomodels.LineBookingScope {
	if info == nil {
		return nil
	}
	for _, r := range info.Roles {
		if slices.Contains(seeAllRoles, r) {
			return nil
		}
	}
	return &repomodels.LineBookingScope{
		UserID:  info.UserID,
		Sales:   slices.Contains(info.Roles, role.Sales),
		Finance: slices.Contains(info.Roles, role.Finance),
	}
}
