package usecase

import (
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/shared/services/linebookingscope"
)

// ownerScope returns what the caller is limited to, or nil when they see every booking.
func ownerScope(info *coremodels.UserRequestInfo) *repomodels.LineBookingScope {
	return linebookingscope.For(info)
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
