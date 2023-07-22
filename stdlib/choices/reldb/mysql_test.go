package reldb

import (
	"testing"
   "fmt"
)

/*
sudo docker run  -e MYSQL_ROOT_PASSWORD=pass -e MYSQL_ROOT_HOST=172.17.0.1 -p 3308:3306 -d mysql
*/

func TestCreateTable(t *testing.T) {

   db := GetMySQL("localhost", "3308")

   conn, err := db.Open("root", "pass", "tester")

   if err != nil {
	   panic(err)
   }

   defer conn.Close()


   err = conn.Exec("CREATE TABLE animals(Name varchar(255), Type varchar(255))")
   if err != nil{
      panic(err)
   }

}


func TestInsert(t *testing.T) {

   db := GetMySQL("localhost", "3308")

   conn, err := db.Open("root", "pass", "tester")

   if err != nil {
	   panic(err)
   }

   defer conn.Close()

   err = conn.Exec("INSERT INTO animals (Name, Type) VALUES (?, ?);", "Leo", "Lion")
   if err != nil{
      panic(err)
   }

}


func TestSelect(t *testing.T) {

   db := GetMySQL("localhost", "3308")

   conn, err := db.Open("root", "pass", "tester")

   if err != nil {
	   panic(err)
   }

   defer conn.Close()

   res, err := conn.Query("SELECT * FROM animals WHERE Name=?", "Leo")
   if err != nil{
      panic(err)
   }

   var animal struct{
      Name string
      Type string
   }

   if res.Next(){
      fmt.Println("NEXTING..")
      err  = res.Scan(&animal.Name, &animal.Type)
      if err != nil {
         panic(err)
      }
   }

   fmt.Println(animal)
}