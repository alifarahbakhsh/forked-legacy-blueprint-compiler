package components

type NoSQLDatabase interface {
	GetDatabase(db_name string) Database
}

type Database interface {
	GetCollection(coll_name string) Collection
}

type Result interface {
	Decode(obj interface{}) error
	All(obj interface{}) error //similar logic to Decode, but for multiple documents
}

type Collection interface {
	DeleteOne(filter string) error
	DeleteMany(filter string) error
	InsertOne(document interface{}) error
	InsertMany(documents []interface{}) error
	FindOne(filter string, projection ...string) (Result, error) //projections should be optional,just like they are in go-mongo and py-mongo. In go-mongo they use an explicit SetProjection method.
	FindMany(filter string, projection ...string) (Result, error) // Result is not a slice -> it is an object we can use to retrieve documents using res.All().
	UpdateOne(filter string, update string) error
	UpdateMany(filter string, update string) error
	ReplaceOne(filter string, replacement interface{}) error
	ReplaceMany(filter string, replacements ...interface{}) error
}

