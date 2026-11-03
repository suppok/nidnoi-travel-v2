package revalidation

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/core/config"
)

type fakeHook struct {
	mu         sync.Mutex
	calls      []time.Time
	failFirst  int
	configured bool
}

func (f *fakeHook) Configured() bool { return f.configured }

func (f *fakeHook) Trigger(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, time.Now())
	if len(f.calls) <= f.failFirst {
		return errors.New("boom")
	}
	return nil
}

func (f *fakeHook) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func start(t *testing.T, hook *fakeHook, quiet, maxWait time.Duration) (*deployHookRevalidator, context.CancelFunc, chan struct{}) {
	t.Helper()
	r := New(hook, zap.NewNop(), config.RevalidationConfig{Quiet: quiet, MaxWait: maxWait}).(*deployHookRevalidator)
	r.retryDelays = []time.Duration{5 * time.Millisecond, 5 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { r.Run(ctx); close(done) }()
	return r, cancel, done
}

func TestDebounceCoalescesBurst(t *testing.T) {
	hook := &fakeHook{configured: true}
	r, cancel, done := start(t, hook, 60*time.Millisecond, time.Second)
	defer func() { cancel(); <-done }()

	for range 5 {
		r.Enqueue(context.Background(), []string{"/a/"})
		time.Sleep(10 * time.Millisecond)
	}
	if hook.count() != 0 {
		t.Fatalf("fired during the burst")
	}
	time.Sleep(150 * time.Millisecond)
	if got := hook.count(); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}

func TestMaxWaitCapsContinuousChanges(t *testing.T) {
	hook := &fakeHook{configured: true}
	r, cancel, done := start(t, hook, 50*time.Millisecond, 120*time.Millisecond)
	defer func() { cancel(); <-done }()

	deadline := time.Now().Add(220 * time.Millisecond)
	for time.Now().Before(deadline) {
		r.Enqueue(context.Background(), []string{"/a/"}) // never quiet for 50ms
		time.Sleep(20 * time.Millisecond)
	}
	if got := hook.count(); got < 1 {
		t.Fatalf("maxWait did not force a rebuild during continuous changes")
	}
}

func TestRetriesThenSucceeds(t *testing.T) {
	hook := &fakeHook{configured: true, failFirst: 2}
	r, cancel, done := start(t, hook, 10*time.Millisecond, time.Second)
	defer func() { cancel(); <-done }()

	r.Enqueue(context.Background(), []string{"/a/"})
	time.Sleep(120 * time.Millisecond)
	if got := hook.count(); got != 3 {
		t.Fatalf("calls = %d, want 3 (2 failures + success)", got)
	}
}

func TestShutdownFlushesPending(t *testing.T) {
	hook := &fakeHook{configured: true}
	r, cancel, done := start(t, hook, time.Hour, time.Hour)
	r.Enqueue(context.Background(), []string{"/a/"})
	cancel()
	<-done
	if got := hook.count(); got != 1 {
		t.Fatalf("calls = %d, want the pending batch flushed once", got)
	}
}

func TestNotConfiguredSkips(t *testing.T) {
	hook := &fakeHook{configured: false}
	r, cancel, done := start(t, hook, 10*time.Millisecond, time.Second)
	r.Enqueue(context.Background(), []string{"/a/"})
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done
	if hook.count() != 0 {
		t.Fatalf("triggered without a configured hook")
	}
	if err := r.RebuildNow(context.Background()); !IsNotConfigured(err) {
		t.Fatalf("RebuildNow err = %v, want ErrNotConfigured", err)
	}
}

func TestRebuildNowClearsPendingAndRequeuesOnFailure(t *testing.T) {
	hook := &fakeHook{configured: true, failFirst: 1}
	r := New(hook, zap.NewNop(), config.RevalidationConfig{Quiet: time.Hour, MaxWait: time.Hour}).(*deployHookRevalidator)

	r.Enqueue(context.Background(), []string{"/a/"})
	if err := r.RebuildNow(context.Background()); err == nil {
		t.Fatal("want the first trigger to fail")
	}
	if _, pending := r.due(); !pending {
		t.Fatal("failed manual rebuild must keep the batch pending")
	}
	if err := r.RebuildNow(context.Background()); err != nil {
		t.Fatalf("second RebuildNow: %v", err)
	}
	if _, pending := r.due(); pending {
		t.Fatal("successful manual rebuild must clear the batch")
	}
}

func TestNextDaily(t *testing.T) {
	cases := map[string]string{
		"2026-11-03T16:00:00Z": "2026-11-03T17:05:00Z", // 23:00 ICT → 00:05 ICT next day
		"2026-11-03T17:05:00Z": "2026-11-04T17:05:00Z", // exactly at 00:05 → tomorrow
		"2026-11-03T17:04:00Z": "2026-11-03T17:05:00Z",
	}
	for in, want := range cases {
		now, _ := time.Parse(time.RFC3339, in)
		if got := nextDaily("00:05", now).UTC().Format(time.RFC3339); got != want {
			t.Errorf("nextDaily(%s) = %s, want %s", in, got, want)
		}
	}
	if !nextDaily("", time.Now()).IsZero() {
		t.Error("empty dailyAt must disable the schedule")
	}
}
