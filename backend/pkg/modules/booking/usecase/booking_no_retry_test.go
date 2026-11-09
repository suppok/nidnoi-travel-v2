package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/modules/booking/models"
	"backend/pkg/registry"
	"backend/pkg/repository"
)

// Fakes: only the methods PublicCreate calls; the embedded interfaces panic on anything else.

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(ctx context.Context, appDB query.AppIDB) error) error {
	return fn(ctx, nil)
}

type fakeTourRepo struct {
	repository.TourProgramRepository
}

func (r fakeTourRepo) WithTx(query.AppIDB) repository.TourProgramRepository { return r }
func (fakeTourRepo) FindPublicStateByCode(context.Context, string) (*domain.TourProgram, error) {
	return &domain.TourProgram{ID: "t1", Code: "NIDN000001", Name: "Tour", State: tourprogramstate.Selling}, nil
}

type fakePeriodRepo struct {
	repository.TourProgramPeriodRepository
}

func (r fakePeriodRepo) WithTx(query.AppIDB) repository.TourProgramPeriodRepository { return r }
func (fakePeriodRepo) FindByID(_ context.Context, id string) (*domain.TourProgramPeriod, error) {
	start := time.Now().AddDate(0, 1, 0)
	return &domain.TourProgramPeriod{ID: id, TourProgramID: "t1", Status: true, StartDate: start, EndDate: start.AddDate(0, 0, 5), Price: 10000, SeatsAvailable: 10}, nil
}

var errTaken = errors.New("duplicate booking_no")

// fakeBookingRepo rejects booking numbers it already holds, like the unique index.
type fakeBookingRepo struct {
	repository.BookingRepository
	taken    map[string]bool
	attempts []string
	failAll  bool
}

func (r *fakeBookingRepo) WithTx(query.AppIDB) repository.BookingRepository { return r }
func (r *fakeBookingRepo) Create(_ context.Context, b *domain.Booking) (*domain.Booking, error) {
	r.attempts = append(r.attempts, b.BookingNo)
	if r.failAll || r.taken[b.BookingNo] {
		return nil, errTaken
	}
	r.taken[b.BookingNo] = true
	return b, nil
}

func newRetryUseCase(t *testing.T, bookings *fakeBookingRepo) *bookingUseCase {
	t.Helper()
	orig := isBookingNoTaken
	isBookingNoTaken = func(err error) bool { return errors.Is(err, errTaken) }
	t.Cleanup(func() { isBookingNoTaken = orig })
	return &bookingUseCase{
		coreRegistry: &registry.CoreRegistry{DBTx: fakeTx{}, Logger: zap.NewNop()},
		repositoryRegistry: &registry.RepositoryRegistry{
			TourProgramRepository:       fakeTourRepo{},
			TourProgramPeriodRepository: fakePeriodRepo{},
			BookingRepository:           bookings,
		},
	}
}

func publicRequest() *models.BookingPublicCreateRequest {
	r := validRequest()
	r.TourCode = "NIDN000001"
	return r
}

func TestPublicCreateRetriesOnTakenNumber(t *testing.T) {
	repo := &fakeBookingRepo{taken: map[string]bool{}}
	u := newRetryUseCase(t, repo)

	// The first two numbers collide as if they already existed.
	calls := 0
	u.repositoryRegistry.BookingRepository = &failingRepo{fakeBookingRepo: repo, failures: 2, calls: &calls}

	res, err := u.PublicCreate(context.Background(), publicRequest())
	if err != nil {
		t.Fatalf("PublicCreate() error = %v, want success after retries", err)
	}
	if calls != 3 {
		t.Fatalf("Create called %d times, want 3 (two collisions, then success)", calls)
	}
	if res.BookingNo != repo.attempts[2] {
		t.Fatalf("returned %q, want the number that was stored (%q)", res.BookingNo, repo.attempts[2])
	}
	if repo.attempts[0] == repo.attempts[1] && repo.attempts[1] == repo.attempts[2] {
		t.Fatal("every attempt reused the same number; want a new number per attempt")
	}
}

func TestPublicCreateGivesUpAfterMaxAttempts(t *testing.T) {
	repo := &fakeBookingRepo{taken: map[string]bool{}, failAll: true}
	u := newRetryUseCase(t, repo)

	if _, err := u.PublicCreate(context.Background(), publicRequest()); err == nil {
		t.Fatal("PublicCreate() with every number taken: want error")
	}
	if len(repo.attempts) != maxNoAttempts {
		t.Fatalf("Create called %d times, want %d", len(repo.attempts), maxNoAttempts)
	}
}

func TestPublicCreateNeverStoresDuplicates(t *testing.T) {
	repo := &fakeBookingRepo{taken: map[string]bool{}}
	u := newRetryUseCase(t, repo)

	for range 2000 {
		if _, err := u.PublicCreate(context.Background(), publicRequest()); err != nil {
			t.Fatalf("PublicCreate() error = %v", err)
		}
	}
	if len(repo.taken) != 2000 {
		t.Fatalf("stored %d distinct numbers for 2000 bookings", len(repo.taken))
	}
}

// failingRepo rejects the first `failures` inserts as taken, then delegates.
type failingRepo struct {
	*fakeBookingRepo
	failures int
	calls    *int
}

func (r *failingRepo) WithTx(query.AppIDB) repository.BookingRepository { return r }
func (r *failingRepo) Create(ctx context.Context, b *domain.Booking) (*domain.Booking, error) {
	*r.calls++
	if *r.calls <= r.failures {
		r.attempts = append(r.attempts, b.BookingNo)
		return nil, errTaken
	}
	return r.fakeBookingRepo.Create(ctx, b)
}
