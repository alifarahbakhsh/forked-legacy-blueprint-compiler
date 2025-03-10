// Blueprint: auto-generated by Blueprint core
package main

import "container5/proc3"
import "sync"
import "log"

func main() {
	var webService *proc3.WebServiceImplHandler
	webService = proc3.GetwebService()
	c := make(chan error, 1)
	wg_done := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		defer wg.Done()
		err := proc3.RunwebService(webService )
		if err != nil {
			c <- err
		}
	}()
	go func(){
		wg.Wait()
		wg_done <- true
	}()
	select {
	case err := <- c:
		log.Fatal(err)
	case <- wg_done:
		log.Println("Success")
	}

}
