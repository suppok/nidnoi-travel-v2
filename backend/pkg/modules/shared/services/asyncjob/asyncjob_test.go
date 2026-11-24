package asyncjob

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func waitDone(t *testing.T, r Runner, id, kind, owner string) *Job {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if job, ok := r.Get(id, kind, owner); ok && job.State != StatePending {
			return job
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("job did not finish")
	return nil
}

func TestRunnerResultAndOwnership(t *testing.T) {
	r := New(time.Minute, zap.NewNop())
	defer r.Shutdown()

	release := make(chan struct{})
	id := r.Start("extract", "u1", func(context.Context) (any, error) { <-release; return "draft", nil })
	if job, ok := r.Get(id, "extract", "u1"); !ok || job.State != StatePending {
		t.Fatalf("pending job = %+v, %v", job, ok)
	}
	if _, ok := r.Get(id, "extract", "u2"); ok {
		t.Fatal("another user must not see the job")
	}
	if _, ok := r.Get(id, "other", "u1"); ok {
		t.Fatal("another kind must not match")
	}
	close(release)
	if job := waitDone(t, r, id, "extract", "u1"); job.State != StateDone || job.Result != "draft" {
		t.Fatalf("job = %+v", job)
	}

	boom := errors.New("boom")
	id = r.Start("extract", "u1", func(context.Context) (any, error) { return nil, boom })
	if job := waitDone(t, r, id, "extract", "u1"); job.State != StateFailed || !errors.Is(job.Err, boom) {
		t.Fatalf("failed job = %+v", job)
	}

	id = r.Start("extract", "u1", func(context.Context) (any, error) { panic("oops") })
	if job := waitDone(t, r, id, "extract", "u1"); job.State != StateFailed {
		t.Fatalf("panicking job = %+v", job)
	}
}

func TestRunnerExpiresFinishedJobs(t *testing.T) {
	r := New(time.Minute, zap.NewNop()).(*runner)
	defer r.Shutdown()
	now := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }

	id := r.Start("extract", "u1", func(context.Context) (any, error) { return 1, nil })
	waitDone(t, r, id, "extract", "u1")
	now = now.Add(2 * time.Minute)
	if _, ok := r.Get(id, "extract", "u1"); ok {
		t.Fatal("job past its TTL must be gone")
	}
}

func TestRunnerShutdownCancelsJobs(t *testing.T) {
	r := New(time.Minute, zap.NewNop())
	id := r.Start("extract", "u1", func(ctx context.Context) (any, error) { <-ctx.Done(); return nil, ctx.Err() })
	r.Shutdown()
	if job := waitDone(t, r, id, "extract", "u1"); !errors.Is(job.Err, context.Canceled) {
		t.Fatalf("job = %+v", job)
	}
}
