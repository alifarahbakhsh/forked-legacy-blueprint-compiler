package reldb

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/stdlib/components"
)

func GetMySQL(addr, port string) *MySqlDB {

	return &MySqlDB{
		addr: addr,
		port: port,
	}
}

type MySqlResult struct {
	underlyingResult *sql.Rows
}

func (mr *MySqlResult) Scan(dest ...interface{}) error {
	return mr.underlyingResult.Scan(dest...)
}

func (mr *MySqlResult) Next() bool {
	return mr.underlyingResult.Next()
}

type MySqlConnection struct {
	conn *sql.DB
}

func (mc *MySqlConnection) Query(query string, args ...interface{}) (components.RelationalDatabaseResult, error) {

	res, err := mc.conn.Query(query, args...)

	if err != nil {
		return nil, err
	}

	return &MySqlResult{
		underlyingResult: res,
	}, nil
}

func (mc *MySqlConnection) Close() error {

	mc.conn.Close()
	return nil
}

func (mc *MySqlConnection) Exec(query string, args ...interface{}) error {

	_, err := mc.conn.Exec(query, args...)

	return err
}

type MySqlDB struct {
	addr string
	port string
}

func (m *MySqlDB) Open(username, password, database string) (components.RelationalDatabaseConnection, error) {

	var err error
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+m.addr+":"+m.port+")/")

	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + database)
	if err != nil {
		return nil, err
	}

	db.Close()

	dbConnection, err := sql.Open("mysql", username+":"+password+"@tcp("+m.addr+":"+m.port+")/"+database)

	if err != nil {
		return nil, err
	}

	return &MySqlConnection{
		conn: dbConnection,
	}, nil
}
