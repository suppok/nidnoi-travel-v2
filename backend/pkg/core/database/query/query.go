package query

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

// AppIDB is *bun.DB or bun.Tx.
type AppIDB = bun.IDB

// DB is a minimal typed access layer over Bun. Extend it only when a use case needs more.
type DB[T any] interface {
	Insert(ctx context.Context, obj *T) (*T, error)                    // RETURNING *
	Update(ctx context.Context, obj *T, columns ...string) (*T, error) // RETURNING *, by primary key
	FindByID(ctx context.Context, id string) (*T, error)               // nil, nil when not found
	NewSelect() *bun.SelectQuery                                       // for custom queries
	NewUpdate() *bun.UpdateQuery                                       // for custom bulk updates
}

type db[T any] struct{ idb AppIDB }

func New[T any](idb AppIDB) DB[T] {
	return &db[T]{idb: idb}
}

func (d *db[T]) Insert(ctx context.Context, obj *T) (*T, error) {
	if _, err := d.idb.NewInsert().Model(obj).Returning("*").Exec(ctx); err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *db[T]) Update(ctx context.Context, obj *T, columns ...string) (*T, error) {
	q := d.idb.NewUpdate().Model(obj).WherePK().Returning("*")
	if len(columns) > 0 {
		q = q.Column(columns...)
	}
	if _, err := q.Exec(ctx); err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *db[T]) FindByID(ctx context.Context, id string) (*T, error) {
	obj := new(T)
	err := d.idb.NewSelect().Model(obj).Where("?TableAlias.id = ?", id).Limit(1).Scan(ctx)
	return NotFoundAsNil(obj, err)
}

func (d *db[T]) NewSelect() *bun.SelectQuery {
	return d.idb.NewSelect().Model((*T)(nil))
}

func (d *db[T]) NewUpdate() *bun.UpdateQuery {
	return d.idb.NewUpdate().Model((*T)(nil))
}

// NotFoundAsNil maps sql.ErrNoRows to (nil, nil), the repository convention for "not found".
func NotFoundAsNil[T any](obj *T, err error) (*T, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// DBTx runs a function inside a transaction.
type DBTx interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, appDB AppIDB) error) error
}

type dbTx struct{ db *bun.DB }

func NewDBTx(db *bun.DB) DBTx {
	return &dbTx{db: db}
}

// WithTx commits when fn returns nil and rolls back otherwise.
func (t *dbTx) WithTx(ctx context.Context, fn func(ctx context.Context, appDB AppIDB) error) error {
	return t.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx, tx)
	})
}
