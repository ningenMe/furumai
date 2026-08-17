// Package cassandra is furumai's Cassandra Stimulus/Observation adapter.
// Its capability shape mirrors adapter/mysql/adapter/postgres (Exec +
// Snapshot returning a DataSet of Row = map[string]any), since Cassandra
// is queried with CQL much like a relational database, despite being a
// NoSQL/wide-column store (see docs/adapter-capability-catalog.md).
package cassandra

import "github.com/gocql/gocql"

// Stimulus runs CQL against a Cassandra cluster. It is used from both
// given and when steps, since Stimulus adapters are shared between them,
// and also serves as the Observation side via Snapshot.
type Stimulus struct {
	Session *gocql.Session
}

// NewStimulus connects to a cluster made up of hosts.
func NewStimulus(hosts ...string) (*Stimulus, error) {
	cluster := gocql.NewCluster(hosts...)
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &Stimulus{Session: session}, nil
}

// Close closes the underlying session.
func (s *Stimulus) Close() { s.Session.Close() }

// Exec runs an arbitrary CQL statement. Use ? placeholders.
func (s *Stimulus) Exec(cql string, args ...any) error {
	return s.Session.Query(cql, args...).Exec()
}

// Row is the full-state Observation for a single CQL result row: column
// name to value. Matchers (Any, Regex, ...) can be substituted for any
// column since map values are any-typed.
type Row map[string]any

// TableQuery scopes one table within a Snapshot call to a WHERE clause
// (the test's namespace/filter, per docs/adapter-capability-catalog.md's
// scoping convention). Table may include a keyspace prefix
// ("keyspace.table"). Use ? placeholders in Where. Where may be empty to
// fetch every row (requires ALLOW FILTERING semantics to not apply, i.e.
// typically only safe on small test tables).
type TableQuery struct {
	Table string
	Where string
	Args  []any
}

// DataSet is the full-state Observation for one or more tables, modeled on
// dbunit's IDataSet: every column of every matching row, per table. Values
// are any-typed (rather than []Row) so a table's rows can be wrapped in
// furumai.AnyOrder, which is typical for Cassandra since row order isn't
// guaranteed unless the query specifies an ORDER BY clause.
//
// There is deliberately no way to build a DataSet from a partial query
// (no column list, no raw CQL): Snapshot always selects every column, so
// "full state" isn't something a caller can accidentally opt out of.
type DataSet map[string]any

// Snapshot fetches the full state of every table in queries (SELECT *,
// each scoped by its own Where/Args) as a single DataSet, for structural
// comparison with furumai.Diff/furumai.ThenEqual.
func (s *Stimulus) Snapshot(queries ...TableQuery) (DataSet, error) {
	result := make(DataSet, len(queries))
	for _, q := range queries {
		rows, err := s.selectAll(q.Table, q.Where, q.Args...)
		if err != nil {
			return nil, err
		}
		result[q.Table] = rows
	}
	return result, nil
}

func (s *Stimulus) selectAll(table, where string, args ...any) ([]Row, error) {
	cql := "SELECT * FROM " + table
	if where != "" {
		cql += " WHERE " + where
	}

	iter := s.Session.Query(cql, args...).Iter()

	var result []Row
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		result = append(result, Row(row))
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return result, nil
}
