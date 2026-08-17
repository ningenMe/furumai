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

// DataSet is the full-state Observation for one or more tables, modeled on
// dbunit's IDataSet: every column of every row, per table (table may
// include a keyspace prefix, "keyspace.table"). Values are any-typed
// (rather than []Row) so a table's rows can be wrapped in
// furumai.AnyOrder, which is typical for Cassandra since row order isn't
// guaranteed unless the query specifies an ORDER BY clause.
//
// There is deliberately no way to build a DataSet from a partial query (no
// column list, no WHERE, no raw CQL): Snapshot always selects every column
// of every row in a table, so "full state" isn't something a caller can
// opt out of, whether by cherry-picking columns or by cherry-picking rows.
// Scoping a test to its own data is a test-isolation concern (e.g.
// TRUNCATE before each test via Exec), not something this interface
// should negotiate away.
type DataSet map[string]any

// Snapshot fetches the full state of every named table (SELECT *, no
// filter) as a single DataSet, for structural comparison with
// furumai.Diff/furumai.ThenEqual.
func (s *Stimulus) Snapshot(tables ...string) (DataSet, error) {
	result := make(DataSet, len(tables))
	for _, table := range tables {
		rows, err := s.selectAll(table)
		if err != nil {
			return nil, err
		}
		result[table] = rows
	}
	return result, nil
}

func (s *Stimulus) selectAll(table string) ([]Row, error) {
	iter := s.Session.Query("SELECT * FROM " + table).Iter()

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
