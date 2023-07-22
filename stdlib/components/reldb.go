package components

type RelationalDB interface{
	Open(username, password, database string) (RelationalDatabaseConnection, error)
}

//* These funcs can be found in either of DB or Conn types from the generic sql driver in Go
//* DB = pooling, Conn = single connection
type RelationalDatabaseConnection interface {
	Query(query string, args ...interface{}) (RelationalDatabaseResult, error)
	Close() error
	Exec(query string, args ...interface{}) error
}

//* Akin to the Result interface from go-sql package
type RelationalDatabaseResult interface{
	//* MySQL actually implements a version of this with a variadic parameter, 
	//* since the `Result` return type of Exec() in sql-driver is just an interface with no precise definition
	//* of the functions that the return type of Exec() has
	Scan(dest ...interface{}) error
	Next() bool
}