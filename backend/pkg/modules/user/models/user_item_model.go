package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// UserItem is the staff user shape returned by every user endpoint (never the password hash).
type UserItem struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Name               string    `json:"name"`
	Roles              []string  `json:"roles"`
	Status             bool      `json:"status"`
	MustChangePassword bool      `json:"mustChangePassword"`
	LastLoginAt        null.Time `json:"lastLoginAt" swaggertype:"string"`
	CreateDate         time.Time `json:"createDate"`
	UpdateDate         time.Time `json:"updateDate"`
}

func NewUserItem(u *domain.AppUser) UserItem {
	return UserItem{
		ID:                 u.ID,
		Email:              u.Email,
		Name:               u.Name,
		Roles:              u.Roles,
		Status:             u.Status,
		MustChangePassword: u.MustChangePassword,
		LastLoginAt:        u.LastLoginAt,
		CreateDate:         u.CreateDate,
		UpdateDate:         u.UpdateDate,
	}
}
