package usecase

import (
	"slices"

	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

// seeAllRoles see every LINE booking. Sales see the bookings they own, finance the bookings sent
// to them on confirm; a user with both roles sees both.
var seeAllRoles = []string{role.SuperAdmin, role.Admin}

// ownerScope returns what the caller is limited to, or nil when they see every booking.
func ownerScope(info *coremodels.UserRequestInfo) *repomodels.LineBookingScope {
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

// visibleTo reports whether the caller may see b. Hidden bookings are reported as NOT_FOUND so
// their existence is not revealed.
func visibleTo(b *domain.LineBooking, info *coremodels.UserRequestInfo) bool {
	s := ownerScope(info)
	if s == nil {
		return true
	}
	return (s.Sales && b.SalesUserID.Valid && b.SalesUserID.String == s.UserID) ||
		(s.Finance && b.FinanceUserID.Valid && b.FinanceUserID.String == s.UserID)
}

// sendsDocuments reports whether the caller may send b's finance documents: its finance user, or a
// see-all user.
func sendsDocuments(b *domain.LineBooking, info *coremodels.UserRequestInfo) bool {
	s := ownerScope(info)
	return s == nil || (s.Finance && b.FinanceUserID.Valid && b.FinanceUserID.String == s.UserID)
}

// ownsAsSales reports whether the caller may act on b as its sales owner (confirm, cancel):
// the owner themselves, or a see-all user.
func ownsAsSales(b *domain.LineBooking, info *coremodels.UserRequestInfo) bool {
	s := ownerScope(info)
	return s == nil || (s.Sales && b.SalesUserID.Valid && b.SalesUserID.String == s.UserID)
}
