package revalidation

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/sitebuild"
	"backend/pkg/core/config"
)

type fakeLog struct {
	mu     sync.Mutex
	builds []Build
}

func (l *fakeLog) Record(b Build) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.builds = append(l.builds, b)
}

func (l *fakeLog) all() []Build {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Build(nil), l.builds...)
}

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
	r := New(hook, &fakeLog{}, zap.NewNop(), config.RevalidationConfig{Quiet: quiet, MaxWait: maxWait}).(*deployHookRevalidator)
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
	builds := r.builds.(*fakeLog).all()
	if len(builds) != 1 || builds[0].Trigger != sitebuild.TriggerAuto || builds[0].Outcome != sitebuild.OutcomeTriggered ||
		len(builds[0].Paths) != 1 || builds[0].Paths[0] != "/a/" {
		t.Fatalf("recorded %+v, want one auto/triggered build for /a/", builds)
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
	if b := r.builds.(*fakeLog).all(); len(b) != 1 || b[0].Attempts != 3 || b[0].Outcome != sitebuild.OutcomeTriggered {
		t.Fatalf("recorded %+v, want one triggered build after 3 attempts", b)
	}
}

func TestGivesUpAndRecordsFailure(t *testing.T) {
	hook := &fakeHook{configured: true, failFirst: 99}
	r, cancel, done := start(t, hook, 10*time.Millisecond, time.Second)
	defer func() { cancel(); <-done }()

	r.Enqueue(context.Background(), []string{"/a/"})
	time.Sleep(120 * time.Millisecond)
	b := r.builds.(*fakeLog).all()
	if len(b) != 1 || b[0].Outcome != sitebuild.OutcomeFailed || b[0].Attempts != 3 || b[0].Error == "" {
		t.Fatalf("recorded %+v, want one failed build after 3 attempts with an error", b)
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
	if err := r.RebuildNow(context.Background(), "user-1"); !IsNotConfigured(err) {
		t.Fatalf("RebuildNow err = %v, want ErrNotConfigured", err)
	}
	b := r.builds.(*fakeLog).all()
	if len(b) != 2 || b[0].Outcome != sitebuild.OutcomeSkipped ||
		b[1].Trigger != sitebuild.TriggerManual || b[1].Outcome != sitebuild.OutcomeSkipped || b[1].ActorID != "user-1" {
		t.Fatalf("recorded %+v, want an auto skip then a manual skip by user-1", b)
	}
}

func TestRebuildNowClearsPendingAndRequeuesOnFailure(t *testing.T) {
	hook := &fakeHook{configured: true, failFirst: 1}
	builds := &fakeLog{}
	r := New(hook, builds, zap.NewNop(), config.RevalidationConfig{Quiet: time.Hour, MaxWait: time.Hour}).(*deployHookRevalidator)

	r.Enqueue(context.Background(), []string{"/a/"})
	if err := r.RebuildNow(context.Background(), "user-1"); err == nil {
		t.Fatal("want the first trigger to fail")
	}
	if _, pending := r.due(); !pending {
		t.Fatal("failed manual rebuild must keep the batch pending")
	}
	if err := r.RebuildNow(context.Background(), "user-1"); err != nil {
		t.Fatalf("second RebuildNow: %v", err)
	}
	if _, pending := r.due(); pending {
		t.Fatal("successful manual rebuild must clear the batch")
	}
	b := builds.all()
	if len(b) != 2 || b[0].Outcome != sitebuild.OutcomeFailed || b[1].Outcome != sitebuild.OutcomeTriggered ||
		b[1].Trigger != sitebuild.TriggerManual || b[1].ActorID != "user-1" || len(b[1].Paths) != 1 {
		t.Fatalf("recorded %+v, want manual failed then manual triggered (covering /a/) by user-1", b)
	}
}

func TestTriggerOf(t *testing.T) {
	if tr, p := triggerOf([]string{DailyPath}); tr != sitebuild.TriggerDaily || len(p) != 0 {
		t.Errorf("daily only: %s %v", tr, p)
	}
	if tr, p := triggerOf([]string{"/", DailyPath, "/tour/X/"}); tr != sitebuild.TriggerAuto || len(p) != 2 {
		t.Errorf("mixed: %s %v, want auto without the daily marker", tr, p)
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
