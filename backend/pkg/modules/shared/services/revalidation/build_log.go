package revalidation

import (
	"context"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/sitebuild"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/repository"
)

// Build is one trigger attempt batch, recorded for the backoffice build history.
type Build struct {
	Trigger  string // sitebuild.Trigger*
	Outcome  string // sitebuild.Outcome*
	Paths    []string
	Attempts int
	Error    string // never contains the hook URL (the deployhook client strips it)
	ActorID  string // user for manual builds; "" = system
}

// BuildLog stores build records. Record never fails the caller: write errors are only logged.
type BuildLog interface {
	Record(b Build)
}

const recordTimeout = 5 * time.Second

type repositoryBuildLog struct {
	repo repository.SiteBuildRepository
	log  *zap.Logger
}

// NewRepositoryBuildLog writes records to site_build. It uses its own context so a record still
// lands while the API is shutting down (the shutdown flush).
func NewRepositoryBuildLog(repo repository.SiteBuildRepository, log *zap.Logger) BuildLog {
	return &repositoryBuildLog{repo: repo, log: log}
}

func (l *repositoryBuildLog) Record(b Build) {
	ctx, cancel := context.WithTimeout(context.Background(), recordTimeout)
	defer cancel()

	actor := b.ActorID
	if actor == "" {
		actor = sitebuild.SystemUserID
	}
	paths := b.Paths
	if paths == nil {
		paths = []string{}
	}
	now := time.Now()
	_, err := l.repo.Create(ctx, &domain.SiteBuild{
		ID:         utils.NewID(),
		Trigger:    b.Trigger,
		Outcome:    b.Outcome,
		Paths:      paths,
		Attempts:   b.Attempts,
		Error:      b.Error,
		Status:     true,
		CreateDate: now,
		UpdateDate: now,
		CreateBy:   actor,
		UpdateBy:   actor,
	})
	if err != nil {
		l.log.Error("record site build", zap.Error(err), zap.String("trigger", b.Trigger), zap.String("outcome", b.Outcome))
	}
}
