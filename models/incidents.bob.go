// Code generated by BobGen psql v0.38.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
	"github.com/stephenafamo/bob/types/pgtypes"
)

// Incident is an object representing the database table.
type Incident struct {
	ID             uuid.UUID           `db:"id,pk" `
	SlackChannelID string              `db:"slack_channel_id" `
	Status         string              `db:"status" `
	Severity       string              `db:"severity" `
	Title          string              `db:"title" `
	Description    sql.Null[string]    `db:"description" `
	StartedBy      string              `db:"started_by" `
	StartedAt      sql.Null[time.Time] `db:"started_at" `
	ResolvedBy     sql.Null[string]    `db:"resolved_by" `
	ResolvedAt     sql.Null[time.Time] `db:"resolved_at" `
	ExportURL      sql.Null[string]    `db:"export_url" `
	LastUpdated    sql.Null[time.Time] `db:"last_updated" `

	R incidentR `db:"-" `
}

// IncidentSlice is an alias for a slice of pointers to Incident.
// This should almost always be used instead of []*Incident.
type IncidentSlice []*Incident

// Incidents contains methods to work with the incidents table
var Incidents = psql.NewTablex[*Incident, IncidentSlice, *IncidentSetter]("", "incidents")

// IncidentsQuery is a query on the incidents table
type IncidentsQuery = *psql.ViewQuery[*Incident, IncidentSlice]

// incidentR is where relationships are stored.
type incidentR struct {
	TimelineEvents TimelineEventSlice // timeline_events.timeline_events_incident_id_fkey
}

type incidentColumnNames struct {
	ID             string
	SlackChannelID string
	Status         string
	Severity       string
	Title          string
	Description    string
	StartedBy      string
	StartedAt      string
	ResolvedBy     string
	ResolvedAt     string
	ExportURL      string
	LastUpdated    string
}

var IncidentColumns = buildIncidentColumns("incidents")

type incidentColumns struct {
	tableAlias     string
	ID             psql.Expression
	SlackChannelID psql.Expression
	Status         psql.Expression
	Severity       psql.Expression
	Title          psql.Expression
	Description    psql.Expression
	StartedBy      psql.Expression
	StartedAt      psql.Expression
	ResolvedBy     psql.Expression
	ResolvedAt     psql.Expression
	ExportURL      psql.Expression
	LastUpdated    psql.Expression
}

func (c incidentColumns) Alias() string {
	return c.tableAlias
}

func (incidentColumns) AliasedAs(alias string) incidentColumns {
	return buildIncidentColumns(alias)
}

func buildIncidentColumns(alias string) incidentColumns {
	return incidentColumns{
		tableAlias:     alias,
		ID:             psql.Quote(alias, "id"),
		SlackChannelID: psql.Quote(alias, "slack_channel_id"),
		Status:         psql.Quote(alias, "status"),
		Severity:       psql.Quote(alias, "severity"),
		Title:          psql.Quote(alias, "title"),
		Description:    psql.Quote(alias, "description"),
		StartedBy:      psql.Quote(alias, "started_by"),
		StartedAt:      psql.Quote(alias, "started_at"),
		ResolvedBy:     psql.Quote(alias, "resolved_by"),
		ResolvedAt:     psql.Quote(alias, "resolved_at"),
		ExportURL:      psql.Quote(alias, "export_url"),
		LastUpdated:    psql.Quote(alias, "last_updated"),
	}
}

type incidentWhere[Q psql.Filterable] struct {
	ID             psql.WhereMod[Q, uuid.UUID]
	SlackChannelID psql.WhereMod[Q, string]
	Status         psql.WhereMod[Q, string]
	Severity       psql.WhereMod[Q, string]
	Title          psql.WhereMod[Q, string]
	Description    psql.WhereNullMod[Q, string]
	StartedBy      psql.WhereMod[Q, string]
	StartedAt      psql.WhereNullMod[Q, time.Time]
	ResolvedBy     psql.WhereNullMod[Q, string]
	ResolvedAt     psql.WhereNullMod[Q, time.Time]
	ExportURL      psql.WhereNullMod[Q, string]
	LastUpdated    psql.WhereNullMod[Q, time.Time]
}

func (incidentWhere[Q]) AliasedAs(alias string) incidentWhere[Q] {
	return buildIncidentWhere[Q](buildIncidentColumns(alias))
}

func buildIncidentWhere[Q psql.Filterable](cols incidentColumns) incidentWhere[Q] {
	return incidentWhere[Q]{
		ID:             psql.Where[Q, uuid.UUID](cols.ID),
		SlackChannelID: psql.Where[Q, string](cols.SlackChannelID),
		Status:         psql.Where[Q, string](cols.Status),
		Severity:       psql.Where[Q, string](cols.Severity),
		Title:          psql.Where[Q, string](cols.Title),
		Description:    psql.WhereNull[Q, string](cols.Description),
		StartedBy:      psql.Where[Q, string](cols.StartedBy),
		StartedAt:      psql.WhereNull[Q, time.Time](cols.StartedAt),
		ResolvedBy:     psql.WhereNull[Q, string](cols.ResolvedBy),
		ResolvedAt:     psql.WhereNull[Q, time.Time](cols.ResolvedAt),
		ExportURL:      psql.WhereNull[Q, string](cols.ExportURL),
		LastUpdated:    psql.WhereNull[Q, time.Time](cols.LastUpdated),
	}
}

var IncidentErrors = &incidentErrors{
	ErrUniqueIncidentsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "incidents",
		columns: []string{"id"},
		s:       "incidents_pkey",
	},
}

type incidentErrors struct {
	ErrUniqueIncidentsPkey *UniqueConstraintError
}

// IncidentSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type IncidentSetter struct {
	ID             *uuid.UUID           `db:"id,pk" `
	SlackChannelID *string              `db:"slack_channel_id" `
	Status         *string              `db:"status" `
	Severity       *string              `db:"severity" `
	Title          *string              `db:"title" `
	Description    *sql.Null[string]    `db:"description" `
	StartedBy      *string              `db:"started_by" `
	StartedAt      *sql.Null[time.Time] `db:"started_at" `
	ResolvedBy     *sql.Null[string]    `db:"resolved_by" `
	ResolvedAt     *sql.Null[time.Time] `db:"resolved_at" `
	ExportURL      *sql.Null[string]    `db:"export_url" `
	LastUpdated    *sql.Null[time.Time] `db:"last_updated" `
}

func (s IncidentSetter) SetColumns() []string {
	vals := make([]string, 0, 12)
	if s.ID != nil {
		vals = append(vals, "id")
	}

	if s.SlackChannelID != nil {
		vals = append(vals, "slack_channel_id")
	}

	if s.Status != nil {
		vals = append(vals, "status")
	}

	if s.Severity != nil {
		vals = append(vals, "severity")
	}

	if s.Title != nil {
		vals = append(vals, "title")
	}

	if s.Description != nil {
		vals = append(vals, "description")
	}

	if s.StartedBy != nil {
		vals = append(vals, "started_by")
	}

	if s.StartedAt != nil {
		vals = append(vals, "started_at")
	}

	if s.ResolvedBy != nil {
		vals = append(vals, "resolved_by")
	}

	if s.ResolvedAt != nil {
		vals = append(vals, "resolved_at")
	}

	if s.ExportURL != nil {
		vals = append(vals, "export_url")
	}

	if s.LastUpdated != nil {
		vals = append(vals, "last_updated")
	}

	return vals
}

func (s IncidentSetter) Overwrite(t *Incident) {
	if s.ID != nil {
		t.ID = *s.ID
	}
	if s.SlackChannelID != nil {
		t.SlackChannelID = *s.SlackChannelID
	}
	if s.Status != nil {
		t.Status = *s.Status
	}
	if s.Severity != nil {
		t.Severity = *s.Severity
	}
	if s.Title != nil {
		t.Title = *s.Title
	}
	if s.Description != nil {
		t.Description = *s.Description
	}
	if s.StartedBy != nil {
		t.StartedBy = *s.StartedBy
	}
	if s.StartedAt != nil {
		t.StartedAt = *s.StartedAt
	}
	if s.ResolvedBy != nil {
		t.ResolvedBy = *s.ResolvedBy
	}
	if s.ResolvedAt != nil {
		t.ResolvedAt = *s.ResolvedAt
	}
	if s.ExportURL != nil {
		t.ExportURL = *s.ExportURL
	}
	if s.LastUpdated != nil {
		t.LastUpdated = *s.LastUpdated
	}
}

func (s *IncidentSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Incidents.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 12)
		if s.ID != nil {
			vals[0] = psql.Arg(*s.ID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.SlackChannelID != nil {
			vals[1] = psql.Arg(*s.SlackChannelID)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.Status != nil {
			vals[2] = psql.Arg(*s.Status)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.Severity != nil {
			vals[3] = psql.Arg(*s.Severity)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.Title != nil {
			vals[4] = psql.Arg(*s.Title)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.Description != nil {
			vals[5] = psql.Arg(*s.Description)
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		if s.StartedBy != nil {
			vals[6] = psql.Arg(*s.StartedBy)
		} else {
			vals[6] = psql.Raw("DEFAULT")
		}

		if s.StartedAt != nil {
			vals[7] = psql.Arg(*s.StartedAt)
		} else {
			vals[7] = psql.Raw("DEFAULT")
		}

		if s.ResolvedBy != nil {
			vals[8] = psql.Arg(*s.ResolvedBy)
		} else {
			vals[8] = psql.Raw("DEFAULT")
		}

		if s.ResolvedAt != nil {
			vals[9] = psql.Arg(*s.ResolvedAt)
		} else {
			vals[9] = psql.Raw("DEFAULT")
		}

		if s.ExportURL != nil {
			vals[10] = psql.Arg(*s.ExportURL)
		} else {
			vals[10] = psql.Raw("DEFAULT")
		}

		if s.LastUpdated != nil {
			vals[11] = psql.Arg(*s.LastUpdated)
		} else {
			vals[11] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s IncidentSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s IncidentSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 12)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.SlackChannelID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "slack_channel_id")...),
			psql.Arg(s.SlackChannelID),
		}})
	}

	if s.Status != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "status")...),
			psql.Arg(s.Status),
		}})
	}

	if s.Severity != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "severity")...),
			psql.Arg(s.Severity),
		}})
	}

	if s.Title != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "title")...),
			psql.Arg(s.Title),
		}})
	}

	if s.Description != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "description")...),
			psql.Arg(s.Description),
		}})
	}

	if s.StartedBy != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "started_by")...),
			psql.Arg(s.StartedBy),
		}})
	}

	if s.StartedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "started_at")...),
			psql.Arg(s.StartedAt),
		}})
	}

	if s.ResolvedBy != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "resolved_by")...),
			psql.Arg(s.ResolvedBy),
		}})
	}

	if s.ResolvedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "resolved_at")...),
			psql.Arg(s.ResolvedAt),
		}})
	}

	if s.ExportURL != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "export_url")...),
			psql.Arg(s.ExportURL),
		}})
	}

	if s.LastUpdated != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "last_updated")...),
			psql.Arg(s.LastUpdated),
		}})
	}

	return exprs
}

// FindIncident retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindIncident(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*Incident, error) {
	if len(cols) == 0 {
		return Incidents.Query(
			SelectWhere.Incidents.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return Incidents.Query(
		SelectWhere.Incidents.ID.EQ(IDPK),
		sm.Columns(Incidents.Columns().Only(cols...)),
	).One(ctx, exec)
}

// IncidentExists checks the presence of a single record by primary key
func IncidentExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return Incidents.Query(
		SelectWhere.Incidents.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after Incident is retrieved from the database
func (o *Incident) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Incidents.AfterSelectHooks.RunHooks(ctx, exec, IncidentSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Incidents.AfterInsertHooks.RunHooks(ctx, exec, IncidentSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Incidents.AfterUpdateHooks.RunHooks(ctx, exec, IncidentSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Incidents.AfterDeleteHooks.RunHooks(ctx, exec, IncidentSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the Incident
func (o *Incident) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *Incident) pkEQ() dialect.Expression {
	return psql.Quote("incidents", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the Incident
func (o *Incident) Update(ctx context.Context, exec bob.Executor, s *IncidentSetter) error {
	v, err := Incidents.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single Incident record with an executor
func (o *Incident) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Incidents.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the Incident using the executor
func (o *Incident) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Incidents.Query(
		SelectWhere.Incidents.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after IncidentSlice is retrieved from the database
func (o IncidentSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Incidents.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Incidents.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Incidents.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Incidents.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o IncidentSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("incidents", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
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
func (o IncidentSlice) copyMatchingRows(from ...*Incident) {
	for i, old := range o {
		for _, new := range from {
			if new.ID != old.ID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o IncidentSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Incidents.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Incident:
				o.copyMatchingRows(retrieved)
			case []*Incident:
				o.copyMatchingRows(retrieved...)
			case IncidentSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Incident or a slice of Incident
				// then run the AfterUpdateHooks on the slice
				_, err = Incidents.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o IncidentSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Incidents.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Incident:
				o.copyMatchingRows(retrieved)
			case []*Incident:
				o.copyMatchingRows(retrieved...)
			case IncidentSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Incident or a slice of Incident
				// then run the AfterDeleteHooks on the slice
				_, err = Incidents.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o IncidentSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals IncidentSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Incidents.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o IncidentSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Incidents.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o IncidentSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Incidents.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type incidentJoins[Q dialect.Joinable] struct {
	typ            string
	TimelineEvents modAs[Q, timelineEventColumns]
}

func (j incidentJoins[Q]) aliasedAs(alias string) incidentJoins[Q] {
	return buildIncidentJoins[Q](buildIncidentColumns(alias), j.typ)
}

func buildIncidentJoins[Q dialect.Joinable](cols incidentColumns, typ string) incidentJoins[Q] {
	return incidentJoins[Q]{
		typ: typ,
		TimelineEvents: modAs[Q, timelineEventColumns]{
			c: TimelineEventColumns,
			f: func(to timelineEventColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, TimelineEvents.Name().As(to.Alias())).On(
						to.IncidentID.EQ(cols.ID),
					))
				}

				return mods
			},
		},
	}
}

// TimelineEvents starts a query for related objects on timeline_events
func (o *Incident) TimelineEvents(mods ...bob.Mod[*dialect.SelectQuery]) TimelineEventsQuery {
	return TimelineEvents.Query(append(mods,
		sm.Where(TimelineEventColumns.IncidentID.EQ(psql.Arg(o.ID))),
	)...)
}

func (os IncidentSlice) TimelineEvents(mods ...bob.Mod[*dialect.SelectQuery]) TimelineEventsQuery {
	pkID := make(pgtypes.Array[uuid.UUID], len(os))
	for i, o := range os {
		pkID[i] = o.ID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkID), "uuid[]")),
	))

	return TimelineEvents.Query(append(mods,
		sm.Where(psql.Group(TimelineEventColumns.IncidentID).OP("IN", PKArgExpr)),
	)...)
}

func (o *Incident) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "TimelineEvents":
		rels, ok := retrieved.(TimelineEventSlice)
		if !ok {
			return fmt.Errorf("incident cannot load %T as %q", retrieved, name)
		}

		o.R.TimelineEvents = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.Incident = o
			}
		}
		return nil
	default:
		return fmt.Errorf("incident has no relationship %q", name)
	}
}

type incidentPreloader struct{}

func buildIncidentPreloader() incidentPreloader {
	return incidentPreloader{}
}

type incidentThenLoader[Q orm.Loadable] struct {
	TimelineEvents func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildIncidentThenLoader[Q orm.Loadable]() incidentThenLoader[Q] {
	type TimelineEventsLoadInterface interface {
		LoadTimelineEvents(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return incidentThenLoader[Q]{
		TimelineEvents: thenLoadBuilder[Q](
			"TimelineEvents",
			func(ctx context.Context, exec bob.Executor, retrieved TimelineEventsLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTimelineEvents(ctx, exec, mods...)
			},
		),
	}
}

// LoadTimelineEvents loads the incident's TimelineEvents into the .R struct
func (o *Incident) LoadTimelineEvents(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.TimelineEvents = nil

	related, err := o.TimelineEvents(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.Incident = o
	}

	o.R.TimelineEvents = related
	return nil
}

// LoadTimelineEvents loads the incident's TimelineEvents into the .R struct
func (os IncidentSlice) LoadTimelineEvents(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	timelineEvents, err := os.TimelineEvents(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.TimelineEvents = nil
	}

	for _, o := range os {
		for _, rel := range timelineEvents {
			if o.ID != rel.IncidentID {
				continue
			}

			rel.R.Incident = o

			o.R.TimelineEvents = append(o.R.TimelineEvents, rel)
		}
	}

	return nil
}

func insertIncidentTimelineEvents0(ctx context.Context, exec bob.Executor, timelineEvents1 []*TimelineEventSetter, incident0 *Incident) (TimelineEventSlice, error) {
	for i := range timelineEvents1 {
		timelineEvents1[i].IncidentID = &incident0.ID
	}

	ret, err := TimelineEvents.Insert(bob.ToMods(timelineEvents1...)).All(ctx, exec)
	if err != nil {
		return ret, fmt.Errorf("insertIncidentTimelineEvents0: %w", err)
	}

	return ret, nil
}

func attachIncidentTimelineEvents0(ctx context.Context, exec bob.Executor, count int, timelineEvents1 TimelineEventSlice, incident0 *Incident) (TimelineEventSlice, error) {
	setter := &TimelineEventSetter{
		IncidentID: &incident0.ID,
	}

	err := timelineEvents1.UpdateAll(ctx, exec, *setter)
	if err != nil {
		return nil, fmt.Errorf("attachIncidentTimelineEvents0: %w", err)
	}

	return timelineEvents1, nil
}

func (incident0 *Incident) InsertTimelineEvents(ctx context.Context, exec bob.Executor, related ...*TimelineEventSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	timelineEvents1, err := insertIncidentTimelineEvents0(ctx, exec, related, incident0)
	if err != nil {
		return err
	}

	incident0.R.TimelineEvents = append(incident0.R.TimelineEvents, timelineEvents1...)

	for _, rel := range timelineEvents1 {
		rel.R.Incident = incident0
	}
	return nil
}

func (incident0 *Incident) AttachTimelineEvents(ctx context.Context, exec bob.Executor, related ...*TimelineEvent) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	timelineEvents1 := TimelineEventSlice(related)

	_, err = attachIncidentTimelineEvents0(ctx, exec, len(related), timelineEvents1, incident0)
	if err != nil {
		return err
	}

	incident0.R.TimelineEvents = append(incident0.R.TimelineEvents, timelineEvents1...)

	for _, rel := range related {
		rel.R.Incident = incident0
	}

	return nil
}
