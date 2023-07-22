package nosqldb

import (
    "testing"
	"reflect"
	"errors"
	"fmt"
)

type someData struct{
	Firstname string
	Lastname string
	Project string
}

func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func TestInsertOne(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	justMe := someData{
		Firstname: "Gerd",
		Lastname: "Alliu",
	}

    err := coll.InsertOne(justMe)

	if err!= nil{
		panic(err)
	}

	//check in db
}

func TestFindOne(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	query := `{"Firstname": "Gerd"}`
    res, err := coll.FindOne(query)

	if err!= nil{
		panic(err)
	}

	var dt someData

	err = res.Decode(&dt)

	if err!= nil{
		panic(err)
	}

	fmt.Println(dt)
	AssertEqual(t, dt.Lastname, "Alliu")
}

func TestUpdateOne(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	query := `{"Firstname": "Gerd"}`
	update := `{"$set": {"Lastname":"Eliah"}}`

    err := coll.UpdateOne(query, update)

	if err!= nil{
		panic(err)
	}
	
    rs, err := coll.FindOne(query)

	if err != nil{
		panic(err)
	}

	var dt someData

	rs.Decode(&dt)
	fmt.Println(dt)
	AssertEqual(t, dt.Lastname, "Eliah")

}


func TestDeleteOne(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	query := `{"Lastname": "Eliah"}`

    err := coll.DeleteOne(query)

	if err!= nil{
		panic(err)
	}
	
    rs, _ := coll.FindOne(query)

	// if err == nil{
	// 	panic(errors.New("Delete failed."))
	// }

	var dt interface{}

	rs.Decode(&dt)
	fmt.Println(dt)
	AssertEqual(t, dt, nil)

}


func TestInsertMany(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	justUs := []interface{}{
		someData{
			Firstname: "Gerd",
			Lastname: "Alliu",
			Project: "Millenial",
		},
		someData{
			Firstname: "Vaastav",
			Lastname: "Anand",
			Project: "Millenial",
		},
	}

    err := coll.InsertMany(justUs)

	if err!= nil{
		panic(err)
	}

	//check in db
}


//TODO, FindMany, DeleteMany

func TestFindMany(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

    res, err := coll.FindMany("")

	if err!= nil{
		panic(err)
	}

	var someData []interface{}

	err = res.All(&someData)

	if err!= nil{
		panic(err)
	}

	fmt.Println(someData)

	if len(someData) != 2{
		panic(errors.New("FindMany failed."))
	}
}

func TestUpdateMany(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	query := `{"Project": "Millenial"}`
	update := `{"$set": {"Project":"GenZ"}}`

    err := coll.UpdateMany(query, update)

	if err!= nil{
		panic(err)
	}
	
	query = `{"Project": "GenZ"}`
    rs, err := coll.FindMany(query)

	if err != nil{
		panic(err)
	}

	var dt []someData

	rs.All(&dt)
	fmt.Println(dt)
	AssertEqual(t, dt[0].Project, "GenZ")
	AssertEqual(t, dt[1].Project, "GenZ")

}


func TestDeleteMany(t *testing.T) {
    
    mng:= GetMongo("localhost", "27017")

    db := mng.GetDatabase("tester")

    coll := db.GetCollection("people")

	query := `{"Project": "GenZ"}`

    err := coll.DeleteMany(query)

	if err!= nil{
		panic(err)
	}

	
    rs, _ := coll.FindMany(query)

	// if err == nil{
	// 	panic(errors.New("Delete failed."))
	// }

	var dt interface{}

	rs.All(&dt)
	fmt.Println(dt)
	AssertEqual(t, dt, nil)
}