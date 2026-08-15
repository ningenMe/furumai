// Package cassandra is furumai's Cassandra Stimulus/Observation adapter.
// Its capability shape mirrors adapter/mysql/adapter/postgres (Exec +
// Query returning Row = map[string]any), since Cassandra is queried with
// CQL much like a relational database, despite being a NoSQL/wide-column
// store (see docs/adapter-capability-catalog.md).
package cassandra

import "github.com/gocql/gocql"

// Stimulus runs CQL against a Cassandra cluster. It is used from both
// given and when steps, since Stimulus adapters are shared between them,
// and also serves as the Observation side via Query.
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

// Query runs cql (use ? placeholders) and returns every matching row as
// the full-state Observation. Compare the result against an expected
// []Row with furumai.Diff/furumai.ThenEqual (wrap the expectation in
// furumai.AnyOrder if row order isn't significant, which is typical for
// Cassandra unless the query specifies an ORDER BY clause).
func (s *Stimulus) Query(cql string, args ...any) ([]Row, error) {
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
