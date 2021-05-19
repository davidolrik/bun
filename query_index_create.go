package bun

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun/internal"
	"github.com/uptrace/bun/schema"
)

type CreateIndexQuery struct {
	whereBaseQuery

	unique       bool
	fulltext     bool
	spatial      bool
	concurrently bool
	ifNotExists  bool

	index   schema.QueryWithArgs
	using   schema.QueryWithArgs
	include []schema.QueryWithArgs
}

func NewCreateIndexQuery(db *DB) *CreateIndexQuery {
	q := &CreateIndexQuery{
		whereBaseQuery: whereBaseQuery{
			baseQuery: baseQuery{
				db:  db,
				dbi: db.DB,
			},
		},
	}
	return q
}

func (q *CreateIndexQuery) DB(db DBI) *CreateIndexQuery {
	q.setDBI(db)
	return q
}

func (q *CreateIndexQuery) Model(model interface{}) *CreateIndexQuery {
	q.setTableModel(model)
	return q
}

func (q *CreateIndexQuery) Unique() *CreateIndexQuery {
	q.unique = true
	return q
}

func (q *CreateIndexQuery) Concurrently() *CreateIndexQuery {
	q.concurrently = true
	return q
}

func (q *CreateIndexQuery) IfNotExists() *CreateIndexQuery {
	q.ifNotExists = true
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Index(query string) *CreateIndexQuery {
	q.index = schema.UnsafeIdent(query)
	return q
}

func (q *CreateIndexQuery) IndexExpr(query string, args ...interface{}) *CreateIndexQuery {
	q.index = schema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Table(tables ...string) *CreateIndexQuery {
	for _, table := range tables {
		q.addTable(schema.UnsafeIdent(table))
	}
	return q
}

func (q *CreateIndexQuery) TableExpr(query string, args ...interface{}) *CreateIndexQuery {
	q.addTable(schema.SafeQuery(query, args))
	return q
}

func (q *CreateIndexQuery) ModelTableExpr(query string, args ...interface{}) *CreateIndexQuery {
	q.modelTable = schema.SafeQuery(query, args)
	return q
}

func (q *CreateIndexQuery) Using(query string, args ...interface{}) *CreateIndexQuery {
	q.using = schema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Column(columns ...string) *CreateIndexQuery {
	for _, column := range columns {
		q.addColumn(schema.UnsafeIdent(column))
	}
	return q
}

func (q *CreateIndexQuery) ColumnExpr(query string, args ...interface{}) *CreateIndexQuery {
	q.addColumn(schema.SafeQuery(query, args))
	return q
}

func (q *CreateIndexQuery) ExcludeColumn(columns ...string) *CreateIndexQuery {
	q.excludeColumn(columns)
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Include(columns ...string) *CreateIndexQuery {
	for _, column := range columns {
		q.include = append(q.include, schema.UnsafeIdent(column))
	}
	return q
}

func (q *CreateIndexQuery) IncludeExpr(query string, args ...interface{}) *CreateIndexQuery {
	q.include = append(q.include, schema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Where(query string, args ...interface{}) *CreateIndexQuery {
	q.addWhere(schema.SafeQueryWithSep(query, args, " AND "))
	return q
}

func (q *CreateIndexQuery) WhereOr(query string, args ...interface{}) *CreateIndexQuery {
	q.addWhere(schema.SafeQueryWithSep(query, args, " OR "))
	return q
}

func (q *CreateIndexQuery) WhereGroup(sep string, fn func(*WhereQuery)) *CreateIndexQuery {
	q.addWhereGroup(sep, fn)
	return q
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	b = append(b, "CREATE "...)

	if q.unique {
		b = append(b, "UNIQUE "...)
	}
	if q.fulltext {
		b = append(b, "FULLTEXT "...)
	}
	if q.spatial {
		b = append(b, "SPATIAL "...)
	}

	b = append(b, "INDEX "...)

	if q.concurrently {
		b = append(b, "CONCURRENTLY "...)
	}
	if q.ifNotExists {
		b = append(b, "IF NOT EXISTS "...)
	}

	b, err = q.index.AppendQuery(fmter, b)
	if err != nil {
		return nil, err
	}

	b = append(b, " ON "...)
	b, err = q.appendFirstTable(fmter, b)
	if err != nil {
		return nil, err
	}

	if !q.using.IsZero() {
		b = append(b, " USING "...)
		b, err = q.using.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	b = append(b, " ("...)
	for i, col := range q.columns {
		if i > 0 {
			b = append(b, ", "...)
		}
		b, err = col.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}
	b = append(b, ')')

	if len(q.include) > 0 {
		b = append(b, " INCLUDE ("...)
		for i, col := range q.include {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = col.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}
		b = append(b, ')')
	}

	b, err = q.appendWhere(fmter, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

//------------------------------------------------------------------------------

func (q *CreateIndexQuery) Exec(
	ctx context.Context, dest ...interface{},
) (res sql.Result, err error) {
	bs := getByteSlice()
	defer putByteSlice(bs)

	queryBytes, err := q.AppendQuery(q.db.fmter, bs.b)
	if err != nil {
		return res, err
	}

	bs.b = queryBytes
	query := internal.String(queryBytes)

	res, err = q.exec(ctx, q, query)
	if err != nil {
		return res, err
	}

	return res, nil
}
