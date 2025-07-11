// Code generated by BobGen psql v0.38.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"io"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/bob/expr"
)

// GooseDBVersion is an object representing the database table.
type GooseDBVersion struct {
	ID        int32     `db:"id,pk" `
	VersionID int64     `db:"version_id" `
	IsApplied bool      `db:"is_applied" `
	Tstamp    time.Time `db:"tstamp" `
}

// GooseDBVersionSlice is an alias for a slice of pointers to GooseDBVersion.
// This should almost always be used instead of []*GooseDBVersion.
type GooseDBVersionSlice []*GooseDBVersion

// GooseDBVersions contains methods to work with the goose_db_version table
var GooseDBVersions = psql.NewTablex[*GooseDBVersion, GooseDBVersionSlice, *GooseDBVersionSetter]("", "goose_db_version")

// GooseDBVersionsQuery is a query on the goose_db_version table
type GooseDBVersionsQuery = *psql.ViewQuery[*GooseDBVersion, GooseDBVersionSlice]

type gooseDBVersionColumnNames struct {
	ID        string
	VersionID string
	IsApplied string
	Tstamp    string
}

var GooseDBVersionColumns = buildGooseDBVersionColumns("goose_db_version")

type gooseDBVersionColumns struct {
	tableAlias string
	ID         psql.Expression
	VersionID  psql.Expression
	IsApplied  psql.Expression
	Tstamp     psql.Expression
}

func (c gooseDBVersionColumns) Alias() string {
	return c.tableAlias
}

func (gooseDBVersionColumns) AliasedAs(alias string) gooseDBVersionColumns {
	return buildGooseDBVersionColumns(alias)
}

func buildGooseDBVersionColumns(alias string) gooseDBVersionColumns {
	return gooseDBVersionColumns{
		tableAlias: alias,
		ID:         psql.Quote(alias, "id"),
		VersionID:  psql.Quote(alias, "version_id"),
		IsApplied:  psql.Quote(alias, "is_applied"),
		Tstamp:     psql.Quote(alias, "tstamp"),
	}
}

type gooseDBVersionWhere[Q psql.Filterable] struct {
	ID        psql.WhereMod[Q, int32]
	VersionID psql.WhereMod[Q, int64]
	IsApplied psql.WhereMod[Q, bool]
	Tstamp    psql.WhereMod[Q, time.Time]
}

func (gooseDBVersionWhere[Q]) AliasedAs(alias string) gooseDBVersionWhere[Q] {
	return buildGooseDBVersionWhere[Q](buildGooseDBVersionColumns(alias))
}

func buildGooseDBVersionWhere[Q psql.Filterable](cols gooseDBVersionColumns) gooseDBVersionWhere[Q] {
	return gooseDBVersionWhere[Q]{
		ID:        psql.Where[Q, int32](cols.ID),
		VersionID: psql.Where[Q, int64](cols.VersionID),
		IsApplied: psql.Where[Q, bool](cols.IsApplied),
		Tstamp:    psql.Where[Q, time.Time](cols.Tstamp),
	}
}

var GooseDBVersionErrors = &gooseDBVersionErrors{
	ErrUniqueGooseDbVersionPkey: &UniqueConstraintError{
		schema:  "",
		table:   "goose_db_version",
		columns: []string{"id"},
		s:       "goose_db_version_pkey",
	},
}

type gooseDBVersionErrors struct {
	ErrUniqueGooseDbVersionPkey *UniqueConstraintError
}

// GooseDBVersionSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type GooseDBVersionSetter struct {
	ID        *int32     `db:"id,pk" `
	VersionID *int64     `db:"version_id" `
	IsApplied *bool      `db:"is_applied" `
	Tstamp    *time.Time `db:"tstamp" `
}

func (s GooseDBVersionSetter) SetColumns() []string {
	vals := make([]string, 0, 4)
	if s.ID != nil {
		vals = append(vals, "id")
	}

	if s.VersionID != nil {
		vals = append(vals, "version_id")
	}

	if s.IsApplied != nil {
		vals = append(vals, "is_applied")
	}

	if s.Tstamp != nil {
		vals = append(vals, "tstamp")
	}

	return vals
}

func (s GooseDBVersionSetter) Overwrite(t *GooseDBVersion) {
	if s.ID != nil {
		t.ID = *s.ID
	}
	if s.VersionID != nil {
		t.VersionID = *s.VersionID
	}
	if s.IsApplied != nil {
		t.IsApplied = *s.IsApplied
	}
	if s.Tstamp != nil {
		t.Tstamp = *s.Tstamp
	}
}

func (s *GooseDBVersionSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return GooseDBVersions.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 4)
		if s.ID != nil {
			vals[0] = psql.Arg(*s.ID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.VersionID != nil {
			vals[1] = psql.Arg(*s.VersionID)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.IsApplied != nil {
			vals[2] = psql.Arg(*s.IsApplied)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.Tstamp != nil {
			vals[3] = psql.Arg(*s.Tstamp)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s GooseDBVersionSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s GooseDBVersionSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 4)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.VersionID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "version_id")...),
			psql.Arg(s.VersionID),
		}})
	}

	if s.IsApplied != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "is_applied")...),
			psql.Arg(s.IsApplied),
		}})
	}

	if s.Tstamp != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "tstamp")...),
			psql.Arg(s.Tstamp),
		}})
	}

	return exprs
}

// FindGooseDBVersion retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindGooseDBVersion(ctx context.Context, exec bob.Executor, IDPK int32, cols ...string) (*GooseDBVersion, error) {
	if len(cols) == 0 {
		return GooseDBVersions.Query(
			SelectWhere.GooseDBVersions.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return GooseDBVersions.Query(
		SelectWhere.GooseDBVersions.ID.EQ(IDPK),
		sm.Columns(GooseDBVersions.Columns().Only(cols...)),
	).One(ctx, exec)
}

// GooseDBVersionExists checks the presence of a single record by primary key
func GooseDBVersionExists(ctx context.Context, exec bob.Executor, IDPK int32) (bool, error) {
	return GooseDBVersions.Query(
		SelectWhere.GooseDBVersions.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after GooseDBVersion is retrieved from the database
func (o *GooseDBVersion) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = GooseDBVersions.AfterSelectHooks.RunHooks(ctx, exec, GooseDBVersionSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = GooseDBVersions.AfterInsertHooks.RunHooks(ctx, exec, GooseDBVersionSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = GooseDBVersions.AfterUpdateHooks.RunHooks(ctx, exec, GooseDBVersionSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = GooseDBVersions.AfterDeleteHooks.RunHooks(ctx, exec, GooseDBVersionSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the GooseDBVersion
func (o *GooseDBVersion) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *GooseDBVersion) pkEQ() dialect.Expression {
	return psql.Quote("goose_db_version", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the GooseDBVersion
func (o *GooseDBVersion) Update(ctx context.Context, exec bob.Executor, s *GooseDBVersionSetter) error {
	v, err := GooseDBVersions.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	*o = *v

	return nil
}

// Delete deletes a single GooseDBVersion record with an executor
func (o *GooseDBVersion) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := GooseDBVersions.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the GooseDBVersion using the executor
func (o *GooseDBVersion) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := GooseDBVersions.Query(
		SelectWhere.GooseDBVersions.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}

	*o = *o2

	return nil
}

// AfterQueryHook is called after GooseDBVersionSlice is retrieved from the database
func (o GooseDBVersionSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = GooseDBVersions.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = GooseDBVersions.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = GooseDBVersions.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = GooseDBVersions.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o GooseDBVersionSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("goose_db_version", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		pkPairs := make([]bob.Expression, len(o))
		for i, row := range o {
			pkPairs[i] = row.primaryKeyVals()
		}
		return bob.ExpressSlice(ctx, w, d, start, pkPairs, "", ", ", "")
	}))
}

// copyMatchingRows finds models in the given slice that have the same primary key
// then it first copies the existing relationships from the old model to the new model
// and then replaces the old model in the slice with the new model
func (o GooseDBVersionSlice) copyMatchingRows(from ...*GooseDBVersion) {
	for i, old := range o {
		for _, new := range from {
			if new.ID != old.ID {
				continue
			}

			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o GooseDBVersionSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return GooseDBVersions.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *GooseDBVersion:
				o.copyMatchingRows(retrieved)
			case []*GooseDBVersion:
				o.copyMatchingRows(retrieved...)
			case GooseDBVersionSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a GooseDBVersion or a slice of GooseDBVersion
				// then run the AfterUpdateHooks on the slice
				_, err = GooseDBVersions.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o GooseDBVersionSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return GooseDBVersions.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *GooseDBVersion:
				o.copyMatchingRows(retrieved)
			case []*GooseDBVersion:
				o.copyMatchingRows(retrieved...)
			case GooseDBVersionSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a GooseDBVersion or a slice of GooseDBVersion
				// then run the AfterDeleteHooks on the slice
				_, err = GooseDBVersions.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o GooseDBVersionSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals GooseDBVersionSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := GooseDBVersions.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o GooseDBVersionSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := GooseDBVersions.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o GooseDBVersionSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := GooseDBVersions.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}
