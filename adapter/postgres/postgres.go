// Package postgres is furumai's PostgreSQL Stimulus/Observation adapter.
// It shares the RDB capability shape with adapter/mysql (see
// docs/adapter-capability-catalog.md); the two are implemented separately
// rather than sharing code so each stays independently mergeable.
package postgres

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Stimulus runs SQL against DB. It is used from both given and when steps,
// since Stimulus adapters are shared between them, and also serves as the
// Observation side via Snapshot.
type Stimulus struct {
	DB *sql.DB
}

// NewStimulus opens a connection pool for dsn (e.g.
// "postgres://user:pass@host:5432/dbname") and verifies it with a ping.
func NewStimulus(dsn string) (*Stimulus, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &Stimulus{DB: db}, nil
}

// Exec runs an arbitrary SQL statement. Use $1, $2, ... placeholders.
func (s *Stimulus) Exec(query string, args ...any) error {
	_, err := s.DB.Exec(query, args...)
	return err
}

// Seed inserts rows into table. Each row is a column name -> value map; rows
// may have different column sets.
func (s *Stimulus) Seed(table string, rows ...map[string]any) error {
	for _, row := range rows {
		cols := make([]string, 0, len(row))
		for col := range row {
			cols = append(cols, col)
		}
		sort.Strings(cols)

		quoted := make([]string, len(cols))
		placeholders := make([]string, len(cols))
		args := make([]any, len(cols))
		for i, col := range cols {
			quoted[i] = `"` + col + `"`
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = row[col]
		}

		query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
			table, strings.Join(quoted, ", "), strings.Join(placeholders, ", "))
		if _, err := s.DB.Exec(query, args...); err != nil {
			return fmt.Errorf("seed %s: %w", table, err)
		}
	}
	return nil
}

// Truncate empties tables.
func (s *Stimulus) Truncate(tables ...string) error {
	for _, table := range tables {
		if _, err := s.DB.Exec(`TRUNCATE TABLE "` + table + `"`); err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}
	return nil
}

// Row is the full-state Observation for a single database row: column name
// to value. Matchers (Any, Regex, ...) can be substituted for any column
// since map values are any-typed.
type Row map[string]any

// TableQuery scopes one table within a Snapshot call to a WHERE clause
// (the test's namespace/filter, per docs/adapter-capability-catalog.md's
// scoping convention). Use $1, $2, ... placeholders in Where. Where may be
// empty to fetch every row.
type TableQuery struct {
	Table string
	Where string
	Args  []any
}

// DataSet is the full-state Observation for one or more tables, modeled on
// dbunit's IDataSet: every column of every matching row, per table. Values
// are any-typed (rather than []Row) so a table's rows can be wrapped in
// furumai.AnyOrder when row order isn't significant.
//
// There is deliberately no way to build a DataSet from a partial query
// (no column list, no raw SQL): Snapshot always selects every column, so
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
			return nil, fmt.Errorf("snapshot %s: %w", q.Table, err)
		}
		result[q.Table] = rows
	}
	return result, nil
}

func (s *Stimulus) selectAll(table, where string, args ...any) ([]Row, error) {
	query := `SELECT * FROM "` + table + `"`
	if where != "" {
		query += " WHERE " + where
	}

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []Row
	for rows.Next() {
		values := make([]any, len(cols))
		scanArgs := make([]any, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		row := make(Row, len(cols))
		for i, col := range cols {
			row[col] = normalize(values[i])
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// normalize converts driver-returned []byte values into string, so query
// results compare cleanly against literal string expectations.
func normalize(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}
