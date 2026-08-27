package mysql

import "errors"

// LoadDataSet reads one or more static fixture files into a DataSet, one
// table per file, for use as Seed input and/or an expected DataSet in
// furumai.ThenEqual (dbunit-style static fixtures instead of inline Go
// literals). Each file's table name is its base name without extension
// (e.g. "testdata/users.csv" -> table "users"). The file format is an
// implementation detail, not part of this function's name or signature.
//
// Not yet implemented; this is the interface shape only.
func LoadDataSet(paths ...string) (DataSet, error) {
	return nil, errors.New("mysql: LoadDataSet not yet implemented")
}
