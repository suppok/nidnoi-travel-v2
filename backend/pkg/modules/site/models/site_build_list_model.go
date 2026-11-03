package models

import (
	"time"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

type SiteBuildListRequest struct {
	Trigger  string `json:"trigger"`  // manual | auto | daily; "" = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type SiteBuildUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SiteBuildItem is one build the backend triggered (not builds from git pushes).
type SiteBuildItem struct {
	ID       string   `json:"id"`
	Trigger  string   `json:"trigger" enums:"manual,auto,daily"`
	Outcome  string   `json:"outcome" enums:"triggered,failed,skipped"`
	Paths    []string `json:"paths"`
	Attempts int      `json:"attempts"`
	Error    string   `json:"error"`
	// TriggeredBy is the user who pressed "rebuild"; null for system builds (auto, daily).
	TriggeredBy *SiteBuildUser `json:"triggeredBy"`
	CreateDate  time.Time      `json:"createDate"`
}

type SiteBuildListResponse struct {
	Items    []SiteBuildItem `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

func NewSiteBuildItem(b *domain.SiteBuild) SiteBuildItem {
	item := SiteBuildItem{
		ID:         b.ID,
		Trigger:    b.Trigger,
		Outcome:    b.Outcome,
		Paths:      b.Paths,
		Attempts:   b.Attempts,
		Error:      b.Error,
		CreateDate: b.CreateDate,
	}
	if item.Paths == nil {
		item.Paths = []string{}
	}
	if b.Creator != nil {
		item.TriggeredBy = &SiteBuildUser{ID: b.Creator.ID, Name: b.Creator.Name}
	}
	return item
}
