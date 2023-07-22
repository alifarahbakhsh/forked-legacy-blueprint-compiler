package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/workload"

	rand2 "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type Stat struct {
	PostID int64
	UserID int64
	Found  bool
}

type PostResponse struct {
	PostID       int64   `json:"Ret0"`
	UserMentions []int64 `json:"Ret1"`
}

type HomeTimelineResponse struct {
	Posts []int64 `json:"Ret0"`
}

var stats []Stat

func printStats(outfile string) {
	stat_strings := []string{}
	for _, stat := range stats {
		stat_strings = append(stat_strings, fmt.Sprintf("%d,%d,%t", stat.PostID, stat.UserID, stat.Found))
	}
	header := "PostID,UserID,Found\n"
	data := header + strings.Join(stat_strings, "\n")
	f, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
}

func get_client() *http.Client {
	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	}
	defaultTransport := *defaultTransportPointer // dereference it to get a copy of the struct that the pointer points to
	defaultTransport.MaxIdleConns = 60000
	defaultTransport.MaxIdleConnsPerHost = 60000
	defaultTransport.MaxConnsPerHost = 10000
	client := &http.Client{Transport: &defaultTransport}
	return client
}

func runCompose(addr string, tput int, finish_chan chan bool, post_chan chan PostResponse) {

	client := get_client()

	tick_every := float64(1e9) / float64(tput)
	var wg sync.WaitGroup
	registry := workload.NewWorkloadRegistry()
	fn := registry.GetGeneratorFunction("sn_ComposePost")
	req_url := addr + "/ComposePost"
	stop_chan := make(chan bool)
	go func() {
		src := rand2.NewSource(0)
		g := distuv.Poisson{100, src}
		timer := time.NewTimer(0 * time.Second)
		next := time.Now()
		for {
			select {
			case <-stop_chan:
				return
			case <-timer.C:
				wg.Add(1)
				go func() {
					defer wg.Done()
					// Make the request
					vals := fn(false)
					user_id, err := strconv.ParseInt(vals.Get("user_id"), 10, 64)
					if err != nil {
						panic(err)
					}
					resp, err := client.PostForm(req_url, vals)
					if err != nil {
						log.Println(err)
						io.Copy(ioutil.Discard, resp.Body)
						resp.Body.Close()
						return
					}
					// Parse and Read Response
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err)
						return
					}
					var post PostResponse
					err = json.Unmarshal(body, &post)
					if err != nil {
						log.Println(err)
						return
					}
					post.UserMentions = []int64{user_id}
					post_chan <- post
				}()
				next = next.Add(time.Duration(g.Rand()*tick_every/100) * time.Nanosecond)
				wait := next.Sub(time.Now())
				timer.Reset(wait)
			}
		}
	}()
	<-finish_chan
	stop_chan <- true
	log.Println("Runner waiting to finish")
	wg.Wait()
	close(post_chan)
	log.Println("Runner over")
}

func runVerifier(addr string, wait time.Duration, finish_chan chan bool, stat_chan chan Stat, post_chan chan PostResponse) {
	client := get_client()
	var wg sync.WaitGroup
	start_post_bytes, _ := json.Marshal(0)
	start_post := string(start_post_bytes)
	stop_post_bytes, _ := json.Marshal(1000)
	stop_post := string(stop_post_bytes)
	req_url := addr + "/ReadUserTimeline"
	wg.Add(1)
	go func() {
		defer wg.Done()
		for post := range post_chan {
			if post.UserMentions == nil {
				continue
			}
			for _, um := range post.UserMentions {
				wg.Add(1)
				go func(um int64, post PostResponse) {
					defer wg.Done()
					time.Sleep(wait)
					// Launch a new request for each user mention
					data := url.Values{}
					user_id_bytes, _ := json.Marshal(um)
					user_id := string(user_id_bytes)
					data.Add("user_id", user_id)
					data.Add("start", start_post)
					data.Add("stop", stop_post)
					resp, err := client.PostForm(req_url, data)
					if err != nil {
						log.Println(err)
						io.Copy(ioutil.Discard, resp.Body)
						resp.Body.Close()
						// Don't count errors
						// Note: not sure about this, maybe we should idk
						return
					}
					statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
					if !statusOK {
						return
					}
					// Parse and Read response
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err)
						// Don't count errors
						// Note: not sure about this, maybe we should idk
						return
					}
					var htr HomeTimelineResponse
					err = json.Unmarshal(body, &htr)
					if err != nil {
						return
					}
					var stat Stat
					stat.UserID = um
					stat.PostID = post.PostID
					stat.Found = false
					//log.Println(um, post.PostID, htr.Posts)
					for _, p := range htr.Posts {
						if p == post.PostID {
							stat.Found = true
							break
						}
					}
					if !stat.Found {
						log.Println(post.PostID, um)
					}
					stat_chan <- stat
				}(um, post)
			}
		}
	}()
	<-finish_chan
	log.Println("Verifier waiting to finish")
	wg.Wait()
	log.Println("Verifier over")
}

func main() {
	addrPtr := flag.String("addr", "localhost", "Path to the configuration file")
	tPutPtr := flag.Int("tput", 100, "Desired Request Rate/Throughput for ComposePost")
	durationPtr := flag.String("duration", "1m", "Duration for which the experiment should run")
	outfilePtr := flag.String("outfile", "results.csv", "File to which the results will be written")
	waitPtr := flag.String("wait", "0s", "Duration for which the verification thread should wait before checking inconsistencies")
	flag.Parse()
	duration, err := time.ParseDuration(*durationPtr)
	if err != nil {
		log.Fatal(err)
	}
	wait, err := time.ParseDuration(*waitPtr)
	if err != nil {
		log.Fatal(err)
	}

	tput := *tPutPtr
	finish_chan := make(chan bool, 2)
	stat_channel := make(chan Stat, *tPutPtr)
	done := make(chan bool)
	post_chan := make(chan PostResponse, tput*10)

	log.Println("Starting Experiment")
	var wg sync.WaitGroup
	// Launch stat collection thread
	go func() {
		i := 0
		for stat := range stat_channel {
			i += 1
			if i%100 == 0 {
				log.Println("Got ", i, " results")
			}
			stats = append(stats, stat)
		}
		close(done)
	}()
	wg.Add(2)
	// Launch compose post thread
	go func() {
		defer wg.Done()
		runCompose(*addrPtr, tput, finish_chan, post_chan)
	}()
	// Launch verification thread
	go func() {
		defer wg.Done()
		runVerifier(*addrPtr, wait, finish_chan, stat_channel, post_chan)
	}()

	time.Sleep(duration)
	log.Println("Sleep Over")
	finish_chan <- true
	finish_chan <- true
	log.Println("Waiting for the end to come")
	wg.Wait()
	close(stat_channel)
	<-done

	log.Println("Writing all stats")
	printStats(*outfilePtr)
}
