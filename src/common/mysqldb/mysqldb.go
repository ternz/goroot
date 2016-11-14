package mysqldb

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const driverName = "mysql"

type MySQLDB struct {
	s  string
	db *sql.DB
}

func NewMySQLDB(s string) (*MySQLDB, error) {
	m := &MySQLDB{s: s}
	db, err := sql.Open(driverName, s)
	if err != nil {
		return nil, err
	}
	m.db = db
	return m, nil
}

func (m *MySQLDB) DB() *sql.DB {
	return m.db
}

type cusField struct {
	dst interface{}
}

func (f *cusField) Scan(src interface{}) error {
	switch s := src.(type) {
	case nil:
		f.dst = nil
	case []byte:
		f.dst = string(s)
	default:
		f.dst = src
	}
	return nil
}

type QueryResult struct {
	D []map[string]interface{}
	E error
}

func (m *MySQLDB) Query(sql string, args []interface{}) *QueryResult {
	r := &QueryResult{}
	rows, err := m.db.Query(sql, args...)
	if err != nil {
		r.E = err
		return r
	}
	columns, err := rows.Columns()
	if err != nil {
		rows.Close()
		r.E = err
		return r
	}
	n := len(columns)
	for rows.Next() {
		dest := make([]interface{}, n)
		for i := 0; i < n; i++ {
			dest[i] = &cusField{}
		}
		if err := rows.Scan(dest...); err != nil {
			rows.Close()
			r.E = err
			return r
		}
		row := make(map[string]interface{})
		for i := 0; i < n; i++ {
			row[columns[i]] = dest[i].(*cusField).dst
		}
		r.D = append(r.D, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		r.E = err
	}
	return r
}

type ExecResult struct {
	LastInsertId int64
	RowsAffected int64
	E            error
}

func (m *MySQLDB) Exec(sql string, args []interface{}) *ExecResult {
	r := &ExecResult{}
	result, err := m.db.Exec(sql, args...)
	if err != nil {
		r.E = err
		return r
	}
	r.LastInsertId, r.E = result.LastInsertId()
	if r.E != nil {
		return r
	}
	r.RowsAffected, r.E = result.RowsAffected()
	return r
}
