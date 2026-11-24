// Package asyncjob runs slow work (AI extraction) in the background so the HTTP request can return at
// once: Cloudflare cuts proxied responses after 100 seconds. The caller polls Get with the job id.
//
// Jobs live in memory, which is correct only because the API runs as exactly one instance. A restart
// loses running and finished jobs; the backoffice then shows "not found" and the user retries.
package asyncjob

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"backend/pkg/core/utils"
)

type State string

const (
	StatePending State = "pending"
	StateDone    State = "done"
	StateFailed  State = "failed"
)

// Job is a snapshot. Result is set when done, Err when failed.
type Job struct {
	ID         string
	Kind       string
	OwnerID    string // only the user who started a job may read it
	State      State
	Result     any
	Err        error
	CreateDate time.Time
	FinishDate time.Time
}

type Runner interface {
	// Start runs fn in the background with its own context (independent of the request, cancelled at
	// shutdown) and returns the job id at once.
	Start(kind, ownerID string, fn func(ctx context.Context) (any, error)) string
	// Get returns the job if it exists, has kind and belongs to ownerID.
	Get(id, kind, ownerID string) (*Job, bool)
	// Shutdown cancels running jobs.
	Shutdown()
}

type runner struct {
	mu     sync.Mutex
	jobs   map[string]*Job
	ttl    time.Duration
	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
	now    func() time.Time
}

// New keeps finished jobs for ttl after they finish (and pending ones for ttl after they start,
// as a backstop).
func New(ttl time.Duration, log *zap.Logger) Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &runner{jobs: map[string]*Job{}, ttl: ttl, log: log, ctx: ctx, cancel: cancel, now: time.Now}
}

func (r *runner) Start(kind, ownerID string, fn func(ctx context.Context) (any, error)) string {
	job := &Job{ID: utils.NewID(), Kind: kind, OwnerID: ownerID, State: StatePending, CreateDate: r.now()}
	r.mu.Lock()
	r.sweepLocked()
	r.jobs[job.ID] = job
	r.mu.Unlock()

	go func() {
		var (
			result any
			err    error
		)
		func() {
			defer func() {
				if p := recover(); p != nil {
					r.log.Error("async job panicked", zap.String("jobId", job.ID), zap.String("kind", kind), zap.Any("panic", p))
					err = errPanic
				}
			}()
			result, err = fn(r.ctx)
		}()
		r.mu.Lock()
		defer r.mu.Unlock()
		job.FinishDate = r.now()
		if err != nil {
			job.State, job.Err = StateFailed, err
		} else {
			job.State, job.Result = StateDone, result
		}
	}()
	return job.ID
}

func (r *runner) Get(id, kind, ownerID string) (*Job, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sweepLocked()
	job, ok := r.jobs[id]
	if !ok || job.Kind != kind || job.OwnerID != ownerID {
		return nil, false
	}
	snapshot := *job
	return &snapshot, true
}

func (r *runner) Shutdown() { r.cancel() }

// sweepLocked drops jobs past their TTL. Called with mu held.
func (r *runner) sweepLocked() {
	now := r.now()
	for id, job := range r.jobs {
		since := job.CreateDate
		if job.State != StatePending {
			since = job.FinishDate
		}
		if now.Sub(since) > r.ttl {
			delete(r.jobs, id)
		}
	}
}

type panicError struct{}

func (panicError) Error() string { return "async job: internal error" }

var errPanic error = panicError{}
