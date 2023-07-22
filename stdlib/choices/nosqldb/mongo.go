package nosqldb

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib/components"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
*	TODOs:
*	1. check bson Unmashal logic
*	2. make sure that the "projection" parameters can also be handled the same way as the `query`
*	3. handle the `_id` cases
 */

//* constructor
func GetMongo(addr, port string) *MongoDB {
	clientOptions := options.Client().ApplyURI("mongodb://" + addr + ":" + port)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		panic(err)
	}
	return &MongoDB{
		client: client,
	}
}

type MongoDB struct {
	client *mongo.Client
}

func (mdb *MongoDB) GetDatabase(dbName string) components.Database {

	dbInstance := mdb.client.Database(dbName)

	return &MongoDatabase{
		db: dbInstance,
	}
}

type MongoDatabase struct {
	db *mongo.Database
}

func (md *MongoDatabase) GetCollection(collectionName string) components.Collection {
	coll := md.db.Collection(collectionName)
	return &MongoCollection{
		collection: coll,
	}
}

type MongoCollection struct {
	collection *mongo.Collection
}

func lower(f interface{}) interface{} {
	switch f := f.(type) {
	case []interface{}:
		for i := range f {
			f[i] = lower(f[i])
		}
		return f
	case map[string]interface{}:
		lf := make(map[string]interface{}, len(f))
		for k, v := range f {
			if k == "$elemMatch" {
				lf[k] = lower(v)
			} else {
				lf[strings.ToLower(k)] = lower(v)
			}
		}
		return lf
	default:
		return f
	}
}

func (mc *MongoCollection) handleFormats(jsonQuery string) (bdoc interface{}, err error) {

	if jsonQuery == "" {
		bdoc = bson.D{}
		return
	}

	var f interface{}
	err = json.Unmarshal([]byte(jsonQuery), &f)
	if err != nil {
		return
	}

	f = lower(f)

	lowerQuery, err := json.Marshal(f)
	if err != nil {
		return
	}
	err = bson.UnmarshalExtJSON(lowerQuery, true, &bdoc)
	return
}

func (mc *MongoCollection) DeleteOne(filter string) error {

	qf, err := mc.handleFormats(filter)
	if err != nil {
		return err
	}

	_, err = mc.collection.DeleteOne(context.TODO(), qf)

	return err

}
func (mc *MongoCollection) DeleteMany(filter string) error {

	qf, err := mc.handleFormats(filter)
	if err != nil {
		return err
	}

	_, err = mc.collection.DeleteMany(context.TODO(), qf)

	return err
}
func (mc *MongoCollection) InsertOne(document interface{}) error {
	_, err := mc.collection.InsertOne(context.TODO(), document)

	return err
}
func (mc *MongoCollection) InsertMany(documents []interface{}) error {
	_, err := mc.collection.InsertMany(context.TODO(), documents)

	return err
}

func (mc *MongoCollection) FindOne(filter string, projection ...string) (components.Result, error) {

	withProjection := false

	if len(projection) > 1 {
		return nil, errors.New("Invalid projection parameter!")
	} else if len(projection) == 1 {
		withProjection = true
	}

	qf, err := mc.handleFormats(filter)
	if err != nil {
		return nil, err
	}

	var singleResult *mongo.SingleResult
	if withProjection {
		prj, err := mc.handleFormats(projection[0])
		if err != nil {
			return nil, err
		}
		opts := options.FindOne().SetProjection(prj)
		singleResult = mc.collection.FindOne(context.TODO(), qf, opts)
	} else {
		singleResult = mc.collection.FindOne(context.TODO(), qf)
	}

	return &MongoResult{
		underlyingResult: singleResult,
	}, nil
}

func (mc *MongoCollection) FindMany(filter string, projection ...string) (components.Result, error) {

	withProjection := false

	if len(projection) > 1 {
		return nil, errors.New("Invalid projection parameter!")
	} else if len(projection) == 1 {
		withProjection = true
	}

	qf, err := mc.handleFormats(filter)
	if err != nil {
		return nil, err
	}

	var cursor *mongo.Cursor
	if withProjection {
		prj, err := mc.handleFormats(projection[0])
		if err != nil {
			return nil, err
		}

		opts := options.Find().SetProjection(prj)
		cursor, err = mc.collection.Find(context.TODO(), qf, opts)
	} else {
		cursor, err = mc.collection.Find(context.TODO(), qf)
	}

	if err != nil {
		return nil, err
	}

	return &MongoResult{
		underlyingResult: cursor,
	}, nil
}

//* not sure about the `update` parameter and its conversion
func (mc *MongoCollection) UpdateOne(filter string, update string) error {
	qf, err := mc.handleFormats(filter)
	if err != nil {
		return err
	}

	up, err := mc.handleFormats(update)
	if err != nil {
		return err
	}

	_, err = mc.collection.UpdateOne(context.TODO(), qf, up)

	return err
}

func (mc *MongoCollection) UpdateMany(filter string, update string) error {
	qf, err := mc.handleFormats(filter)
	if err != nil {
		return err
	}

	up, err := mc.handleFormats(update)
	if err != nil {
		return err
	}

	_, err = mc.collection.UpdateMany(context.TODO(), qf, up)

	return err

}

func (mc *MongoCollection) ReplaceOne(filter string, replacement interface{}) error {
	qf, err := mc.handleFormats(filter)
	if err != nil {
		return err
	}

	_, err = mc.collection.ReplaceOne(context.TODO(), qf, replacement)

	return err
}

func (mc *MongoCollection) ReplaceMany(filter string, replacements ...interface{}) error {
	return errors.New("ReplaceMany not implemented")
}

type MongoResult struct {
	underlyingResult interface{}
}

func (mr *MongoResult) Decode(obj interface{}) error {

	//add other types of results from mongo that have a Decode method here
	switch v := mr.underlyingResult.(type) {
	case *mongo.SingleResult:
		return v.Decode(obj)
	default:
		return errors.New("Result has no decode method")
	}
}

func (mr *MongoResult) All(objs interface{}) error {
	//add other types of results from mongo that are Cursors here
	switch v := mr.underlyingResult.(type) {
	case *mongo.Cursor:
		return v.All(context.TODO(), objs)
	default:
		return errors.New("Result does not return a Cursor")
	}
}
