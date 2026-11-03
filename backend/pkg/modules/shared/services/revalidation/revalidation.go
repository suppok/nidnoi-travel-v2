// Package revalidation rebuilds the static customer site after published content changes.
// Use cases call Enqueue after their transaction commits; a single in-process worker (Run) debounces
// the changes and fires the Cloudflare deploy hook.
package revalidation

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/sitebuild"
	"backend/pkg/core/config"
	"backend/pkg/core/utils"
	"backend/pkg/modules/shared/client/deployhook"
)

// ErrNotConfigured: CLOUDFLARE_DEPLOY_HOOK_URL is not set, so nothing can be rebuilt.
var ErrNotConfigured = deployhook.ErrNotConfigured

// DailyPath marks the scheduled daily rebuild in logs.
const DailyPath = "(daily)"

type Revalidator interface {
	// Enqueue records changed customer-site paths. Non-blocking; never fails (paths are only logged).
	Enqueue(ctx context.Context, paths []string)
	// RebuildNow fires the hook immediately, bypassing the debounce (manual "rebuild" button).
	// actorID is recorded in the build history.
	RebuildNow(ctx context.Context, actorID string) error
}

// Worker is the Revalidator plus its background loop, started once from app.go.
type Worker interface {
	Revalidator
	// Run debounces and fires until ctx is cancelled, then flushes pending work once.
	Run(ctx context.Context)
}

const shutdownFlushTimeout = 5 * time.Second

type deployHookRevalidator struct {
	hook        deployhook.DeployHook
	builds      BuildLog
	log         *zap.Logger
	quiet       time.Duration
	maxWait     time.Duration
	dailyAt     string // HH:MM Thailand time, "" = off
	retryDelays []time.Duration

	mu      sync.Mutex
	pending map[string]struct{}
	first   time.Time // first enqueue of the pending batch
	last    time.Time // latest enqueue of the pending batch
	wake    chan struct{}
}

// New builds the deploy-hook revalidator. Config is validated at startup. Every trigger (and skip)
// is recorded in builds.
func New(hook deployhook.DeployHook, builds BuildLog, log *zap.Logger, cfg config.RevalidationConfig) Worker {
	return &deployHookRevalidator{
		hook:        hook,
		builds:      builds,
		log:         log,
		quiet:       cfg.Quiet,
		maxWait:     cfg.MaxWait,
		dailyAt:     cfg.DailyAt,
		retryDelays: []time.Duration{time.Second, 5 * time.Second, 30 * time.Second},
		pending:     map[string]struct{}{},
		wake:        make(chan struct{}, 1),
	}
}

func (r *deployHookRevalidator) Enqueue(_ context.Context, paths []string) {
	if len(paths) == 0 {
		return
	}
	r.add(paths)
	select {
	case r.wake <- struct{}{}:
	default: // the loop is already due to re-check
	}
}

func (r *deployHookRevalidator) add(paths []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if len(r.pending) == 0 {
		r.first = now
	}
	r.last = now
	for _, p := range paths {
		r.pending[p] = struct{}{}
	}
}

// due reports when the pending batch should fire: quiet after the last change, capped at maxWait
// after the first.
func (r *deployHookRevalidator) due() (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.pending) == 0 {
		return time.Time{}, false
	}
	at := r.last.Add(r.quiet)
	if capAt := r.first.Add(r.maxWait); capAt.Before(at) {
		at = capAt
	}
	return at, true
}

// take empties the pending batch and returns its paths, sorted for stable logs.
func (r *deployHookRevalidator) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	paths := make([]string, 0, len(r.pending))
	for p := range r.pending {
		paths = append(paths, p)
	}
	r.pending = map[string]struct{}{}
	slices.Sort(paths)
	return paths
}

func (r *deployHookRevalidator) Run(ctx context.Context) {
	daily := nextDaily(r.dailyAt, time.Now())
	for {
		wait := time.Hour
		if at, ok := r.due(); ok {
			wait = time.Until(at)
		}
		if !daily.IsZero() {
			wait = min(wait, time.Until(daily))
		}

		select {
		case <-ctx.Done():
			r.flush()
			return
		case <-r.wake:
			continue // re-compute the deadline
		case <-time.After(max(wait, 0)):
		}

		now := time.Now()
		if !daily.IsZero() && !now.Before(daily) {
			r.add([]string{DailyPath})
			daily = nextDaily(r.dailyAt, now)
		}
		if at, ok := r.due(); ok && !now.Before(at) {
			r.fire(ctx, r.take())
		}
	}
}

// triggerOf labels a debounced batch: daily when it holds only the scheduled marker, auto otherwise.
// The marker itself is dropped from the recorded paths.
func triggerOf(paths []string) (string, []string) {
	content := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != DailyPath {
			content = append(content, p)
		}
	}
	if len(content) == 0 {
		return sitebuild.TriggerDaily, content
	}
	return sitebuild.TriggerAuto, content
}

// fire triggers the hook, retrying with backoff. A final failure is logged with the paths it carried.
func (r *deployHookRevalidator) fire(ctx context.Context, batch []string) {
	trigger, paths := triggerOf(batch)
	if !r.hook.Configured() {
		r.log.Info("site rebuild skipped: CLOUDFLARE_DEPLOY_HOOK_URL is not set", zap.String("trigger", trigger), zap.Strings("paths", paths))
		r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeSkipped, Paths: paths})
		return
	}
	var err error
	attempts := 0
	for {
		attempts++
		if err = r.hook.Trigger(ctx); err == nil {
			r.log.Info("site rebuild triggered", zap.String("trigger", trigger), zap.Strings("paths", paths), zap.Int("attempt", attempts))
			r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeTriggered, Paths: paths, Attempts: attempts})
			return
		}
		if attempts > len(r.retryDelays) || ctx.Err() != nil {
			break
		}
		delay := r.retryDelays[attempts-1]
		r.log.Warn("site rebuild trigger failed, retrying", zap.Error(err), zap.Duration("in", delay))
		select {
		case <-ctx.Done():
		case <-time.After(delay):
		}
	}
	r.log.Error("site rebuild trigger failed", zap.Error(err), zap.String("trigger", trigger), zap.Strings("paths", paths))
	r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeFailed, Paths: paths, Attempts: attempts, Error: err.Error()})
}

// flush fires a still-pending batch once on shutdown, so a redeploy right after an edit loses nothing.
func (r *deployHookRevalidator) flush() {
	if _, ok := r.due(); !ok {
		return
	}
	trigger, paths := triggerOf(r.take())
	if !r.hook.Configured() {
		r.log.Info("site rebuild skipped on shutdown: CLOUDFLARE_DEPLOY_HOOK_URL is not set", zap.Strings("paths", paths))
		r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeSkipped, Paths: paths})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownFlushTimeout)
	defer cancel()
	if err := r.hook.Trigger(ctx); err != nil {
		r.log.Error("site rebuild trigger failed on shutdown", zap.Error(err), zap.Strings("paths", paths))
		r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeFailed, Paths: paths, Attempts: 1, Error: err.Error()})
		return
	}
	r.log.Info("site rebuild triggered on shutdown", zap.Strings("paths", paths))
	r.builds.Record(Build{Trigger: trigger, Outcome: sitebuild.OutcomeTriggered, Paths: paths, Attempts: 1})
}

func (r *deployHookRevalidator) RebuildNow(ctx context.Context, actorID string) error {
	if !r.hook.Configured() {
		r.builds.Record(Build{Trigger: sitebuild.TriggerManual, Outcome: sitebuild.OutcomeSkipped, ActorID: actorID})
		return ErrNotConfigured
	}
	// The manual build covers everything pending; put it back if the trigger fails.
	batch := r.take()
	_, paths := triggerOf(batch)
	if err := r.hook.Trigger(ctx); err != nil {
		if len(batch) > 0 {
			r.Enqueue(ctx, batch)
		}
		r.builds.Record(Build{Trigger: sitebuild.TriggerManual, Outcome: sitebuild.OutcomeFailed, Paths: paths, Attempts: 1, Error: err.Error(), ActorID: actorID})
		return err
	}
	r.log.Info("site rebuild triggered manually", zap.Strings("pending", paths))
	r.builds.Record(Build{Trigger: sitebuild.TriggerManual, Outcome: sitebuild.OutcomeTriggered, Paths: paths, Attempts: 1, ActorID: actorID})
	return nil
}

// nextDaily returns the next HH:MM (Thailand time) strictly after now; zero when hhmm is "".
func nextDaily(hhmm string, now time.Time) time.Time {
	if hhmm == "" {
		return time.Time{}
	}
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}
	}
	local := now.In(utils.Thailand)
	at := time.Date(local.Year(), local.Month(), local.Day(), t.Hour(), t.Minute(), 0, 0, utils.Thailand)
	if !at.After(now) {
		at = at.AddDate(0, 0, 1)
	}
	return at
}

// IsNotConfigured reports whether err means no deploy hook is set.
func IsNotConfigured(err error) bool { return errors.Is(err, ErrNotConfigured) }
