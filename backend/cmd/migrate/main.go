// Command migrate applies SQL migrations and bootstraps the first user (the super admin).
//
//	go run ./cmd/migrate up | down | status | create <description>
//	go run ./cmd/migrate create-super-admin <email> <name>
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"

	"backend/pkg/constants/role"
	"backend/pkg/core/config"
	"backend/pkg/core/database"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/services/password"
	"backend/pkg/repository"
)

const (
	configPath    = "config/config_api.yml"
	migrationsDir = "migrations"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	ctx := context.Background()

	if os.Args[1] == "create" {
		if len(os.Args) < 3 {
			usage()
		}
		check(create(ctx, strings.Join(os.Args[2:], "_")))
		return
	}

	cfg, err := config.LoadForMigrate(configPath)
	check(err)
	db, err := database.Open(ctx, cfg)
	check(err)
	defer db.Close()

	switch os.Args[1] {
	case "up":
		check(up(ctx, db))
	case "down":
		check(down(ctx, db))
	case "status":
		check(status(ctx, db))
	case "create-super-admin":
		if len(os.Args) < 4 {
			usage()
		}
		check(createSuperAdmin(ctx, db, cfg.Auth.BcryptCost, os.Args[2], strings.Join(os.Args[3:], " ")))
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrate up | down | status | create <description> | create-super-admin <email> <name>")
	os.Exit(2)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func newMigrator(db *bun.DB) (*migrate.Migrator, error) {
	migrations := migrate.NewMigrations(migrate.WithMigrationsDirectory(migrationsDir))
	if err := migrations.Discover(os.DirFS(migrationsDir)); err != nil {
		return nil, err
	}
	return migrate.NewMigrator(db, migrations), nil
}

func up(ctx context.Context, db *bun.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	if err := m.Init(ctx); err != nil {
		return err
	}
	if err := m.Lock(ctx); err != nil {
		return err
	}
	defer m.Unlock(ctx) //nolint:errcheck

	group, err := m.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Println("no new migrations")
		return nil
	}
	fmt.Printf("migrated to %s\n", group)
	return nil
}

func down(ctx context.Context, db *bun.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	// Creates the bookkeeping tables on a fresh database (no-op otherwise).
	if err := m.Init(ctx); err != nil {
		return err
	}
	if err := m.Lock(ctx); err != nil {
		return err
	}
	defer m.Unlock(ctx) //nolint:errcheck

	group, err := m.Rollback(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Println("no groups to roll back")
		return nil
	}
	fmt.Printf("rolled back %s\n", group)
	return nil
}

func status(ctx context.Context, db *bun.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	// Creates the bookkeeping tables on a fresh database (no-op otherwise).
	if err := m.Init(ctx); err != nil {
		return err
	}
	ms, err := m.MigrationsWithStatus(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("migrations: %s\n", ms)
	fmt.Printf("unapplied:  %s\n", ms.Unapplied())
	fmt.Printf("last group: %s\n", ms.LastGroup())
	return nil
}

// create writes a transactional up/down SQL pair; it needs no database.
func create(ctx context.Context, description string) error {
	migrations := migrate.NewMigrations(migrate.WithMigrationsDirectory(migrationsDir))
	files, err := migrate.NewMigrator(nil, migrations).CreateTxSQLMigrations(ctx, description)
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Printf("created %s\n", f.Path)
	}
	return nil
}

// createSuperAdmin bootstraps the first user with the super_admin role and a random temporary
// password, printed once; it must be changed at first login. Refuses if a super admin or the email
// already exists.
func createSuperAdmin(ctx context.Context, db *bun.DB, bcryptCost int, email, name string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if !strings.Contains(email, "@") || name == "" {
		return errors.New("create-super-admin: a valid email and a name are required")
	}

	passwords, err := password.NewService(bcryptCost)
	if err != nil {
		return err
	}
	temporary, err := passwords.GenerateTemporary()
	if err != nil {
		return err
	}
	hash, err := passwords.Hash(temporary)
	if err != nil {
		return err
	}

	users := repository.NewAppUserRepository(db)
	exists, err := users.ExistsWithRole(ctx, role.SuperAdmin)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("create-super-admin: a super admin already exists")
	}
	existing, err := users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("create-super-admin: %s already exists", email)
	}

	id := utils.NewID()
	now := time.Now()
	_, err = users.Create(ctx, &domain.AppUser{
		ID:                 id,
		Email:              email,
		Name:               name,
		PasswordHash:       hash,
		Roles:              []string{role.SuperAdmin},
		MustChangePassword: true,
		Status:             true,
		CreateDate:         now,
		UpdateDate:         now,
		CreateBy:           id,
		UpdateBy:           id,
	})
	if err != nil {
		return err
	}

	fmt.Printf("super admin created: %s\ntemporary password (shown once, must be changed at first login): %s\n", email, temporary)
	return nil
}
