package mysql

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
)

// LoadCSV reads a dbunit-style flat CSV fixture: the first row is column
// names, each following row is one Row. This lets Seed input and expected
// DataSet values live as static fixture files instead of inline literals.
//
// Cell values are type-inferred (int64, then float64, then string as a
// fallback) so a fixture column compares cleanly against an INT/DECIMAL
// database column without the caller converting anything by hand. This is
// a heuristic, not schema-aware: a value like "007" parses as 7, losing
// the leading zero. Use a matcher (e.g. Regex) in the expected value for
// columns where that matters.
func LoadCSV(path string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	var rows []Row
	for {
		record, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		row := make(Row, len(header))
		for i, col := range header {
			row[col] = inferType(record[i])
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func inferType(s string) any {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}
