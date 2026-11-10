package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/config"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	tourprogramusecase "backend/pkg/modules/tourprogram/usecase"
)

// Mock tour programs for local/dev, taken from the legacy nidnoitravel.com site (multi-country
// tours, real names, prices and departures; seats are made up). Never runs on prd.
//
//	go run ./cmd/migrate seed-mock-tours          # add the tours that are not there yet
//	go run ./cmd/migrate seed-mock-tours remove   # delete every mock tour (and bookings on them)
//
// Seeded rows are marked by internal_note = mockTourNote; wholesale_tour_code keeps the legacy code.
// The customer site shows them after the next rebuild ("อัปเดตหน้าเว็บตอนนี้" in the backoffice).

//go:embed seeddata/mock_tours.json
var mockToursJSON []byte

const mockTourNote = "[mock-seed] ข้อมูลทดสอบจาก nidnoitravel.com — ลบได้ด้วย task seed-mock-tours -- remove"

type mockPeriod struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Price  int    `json:"price"`
	Before int    `json:"before"` // price before discount; 0 = none
	Single int    `json:"single"` // legacy single-room price; supplement = single - price
}

type mockTour struct {
	LegacyCode string       `json:"legacyCode"`
	Name       string       `json:"name"`
	Countries  []string     `json:"countries"` // country slugs, main country first
	Cities     []string     `json:"cities"`    // city slugs
	Days       int          `json:"days"`
	Nights     int          `json:"nights"`
	Airline    string       `json:"airline"` // airline code; "" = none
	Highlight  string       `json:"highlight"`
	Periods    []mockPeriod `json:"periods"`
}

// mockSeats gives the period table every state to show: plenty, a few left, sold out.
func mockSeats(tourIndex, periodIndex int) (total, available int) {
	switch (tourIndex + periodIndex) % 7 {
	case 3:
		return 25, 3
	case 5:
		return 25, 0
	default:
		return 30, 12 + (tourIndex*5+periodIndex*3)%15
	}
}

func seedMockTours(ctx context.Context, db *bun.DB, cfg *config.Config, remove bool) error {
	if cfg.AppEnv != config.EnvLocal && cfg.AppEnv != config.EnvDev {
		return fmt.Errorf("seed-mock-tours: refusing to run with APP_ENV=%q (local or dev only)", cfg.AppEnv)
	}
	if remove {
		return removeMockTours(ctx, db)
	}

	var tours []mockTour
	if err := json.Unmarshal(mockToursJSON, &tours); err != nil {
		return fmt.Errorf("seed-mock-tours: parse data: %w", err)
	}

	var wholesaleID string
	if err := db.NewSelect().Table("wholesale").Column("id").Where("status").OrderExpr("code").Limit(1).Scan(ctx, &wholesaleID); err != nil {
		return fmt.Errorf("seed-mock-tours: no active wholesale: %w", err)
	}

	added, skipped := 0, 0
	for i, t := range tours {
		exists, err := db.NewSelect().Table("tour_program").
			Where("internal_note = ?", mockTourNote).Where("wholesale_tour_code = ?", t.LegacyCode).Exists(ctx)
		if err != nil {
			return err
		}
		if exists {
			skipped++
			continue
		}
		code, err := insertMockTour(ctx, db, cfg, wholesaleID, i, t)
		if err != nil {
			return fmt.Errorf("seed-mock-tours: %s: %w", t.LegacyCode, err)
		}
		log.Printf("added %s (legacy %s) %s", code, t.LegacyCode, t.Name)
		added++
	}
	log.Printf("seed-mock-tours: %d added, %d already there. Rebuild the customer site to show them.", added, skipped)
	return nil
}

func lookupIDs(ctx context.Context, tx bun.Tx, table, column string, values []string) ([]string, error) {
	ids := make([]string, len(values))
	for i, v := range values {
		if err := tx.NewSelect().Table(table).Column("id").Where("? = ?", bun.Ident(column), v).Limit(1).Scan(ctx, &ids[i]); err != nil {
			return nil, fmt.Errorf("%s %q not found: %w", table, v, err)
		}
	}
	return ids, nil
}

func insertMockTour(ctx context.Context, db *bun.DB, cfg *config.Config, wholesaleID string, index int, t mockTour) (string, error) {
	var code string
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		countryIDs, err := lookupIDs(ctx, tx, "country", "slug", t.Countries)
		if err != nil {
			return err
		}
		cityIDs, err := lookupIDs(ctx, tx, "city", "slug", t.Cities)
		if err != nil {
			return err
		}
		airlineID := null.String{}
		if t.Airline != "" {
			ids, err := lookupIDs(ctx, tx, "airline", "code", []string{t.Airline})
			if err != nil {
				return err
			}
			airlineID = null.StringFrom(ids[0])
		}

		var number int64
		if err := tx.NewSelect().ColumnExpr("nextval('tour_code_seq')").Scan(ctx, &number); err != nil {
			return err
		}
		code = tourprogramusecase.FormatTourCode(cfg.TourCode.Prefix, number)

		now := time.Now()
		system := "00000000-0000-0000-0000-000000000000"
		tour := &domain.TourProgram{
			ID:                utils.NewID(),
			Code:              code,
			Name:              t.Name,
			WholesaleID:       wholesaleID,
			WholesaleTourCode: t.LegacyCode,
			State:             tourprogramstate.Selling,
			Days:              t.Days,
			Nights:            t.Nights,
			AirlineID:         airlineID,
			Highlight:         t.Highlight,
			Description:       fmt.Sprintf("%s %d วัน %d คืน\n\nไฮไลต์ของทริป: %s", t.Name, t.Days, t.Nights, t.Highlight),
			InternalNote:      mockTourNote,
			Status:            true,
			CreateDate:        now,
			UpdateDate:        now,
			CreateBy:          system,
			UpdateBy:          system,
		}
		if _, err := tx.NewInsert().Model(tour).Exec(ctx); err != nil {
			return err
		}

		for i, id := range countryIDs {
			link := &domain.TourProgramCountry{TourProgramID: tour.ID, CountryID: id, SortOrder: i}
			if _, err := tx.NewInsert().Model(link).Exec(ctx); err != nil {
				return err
			}
		}
		for i, id := range cityIDs {
			link := &domain.TourProgramCity{TourProgramID: tour.ID, CityID: id, SortOrder: i}
			if _, err := tx.NewInsert().Model(link).Exec(ctx); err != nil {
				return err
			}
		}

		for i, p := range t.Periods {
			start, err := time.Parse(time.DateOnly, p.Start)
			if err != nil {
				return err
			}
			end, err := time.Parse(time.DateOnly, p.End)
			if err != nil {
				return err
			}
			total, available := mockSeats(index, i)
			period := &domain.TourProgramPeriod{
				ID:             utils.NewID(),
				TourProgramID:  tour.ID,
				StartDate:      start,
				EndDate:        end,
				Price:          p.Price,
				SeatsTotal:     total,
				SeatsAvailable: available,
				Status:         true,
				CreateDate:     now,
				UpdateDate:     now,
				CreateBy:       system,
				UpdateBy:       system,
			}
			if p.Before > p.Price {
				period.PriceBeforeDiscount = null.IntFrom(int64(p.Before))
			}
			if p.Single > p.Price {
				period.SingleSupplement = null.IntFrom(int64(p.Single - p.Price))
			}
			if _, err := tx.NewInsert().Model(period).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	return code, err
}

func removeMockTours(ctx context.Context, db *bun.DB) error {
	var ids []string
	if err := db.NewSelect().Table("tour_program").Column("id").Where("internal_note = ?", mockTourNote).Scan(ctx, &ids); err != nil {
		return err
	}
	if len(ids) == 0 {
		log.Print("seed-mock-tours remove: no mock tours found")
		return nil
	}
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Children first; bookings made on mock tours are test data too.
		for _, table := range []string{"booking", "tour_program_period", "tour_program_image", "tour_program_city", "tour_program_country"} {
			res, err := tx.NewDelete().Table(table).Where("tour_program_id IN (?)", bun.In(ids)).Exec(ctx)
			if err != nil {
				return fmt.Errorf("delete %s: %w", table, err)
			}
			n, _ := res.RowsAffected()
			log.Printf("deleted %d %s rows", n, table)
		}
		res, err := tx.NewDelete().Table("tour_program").Where("id IN (?)", bun.In(ids)).Exec(ctx)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != int64(len(ids)) {
			return errors.New("seed-mock-tours remove: tour count changed during delete")
		}
		log.Printf("seed-mock-tours remove: deleted %d mock tours. Rebuild the customer site.", n)
		return nil
	})
}
