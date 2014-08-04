package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

// Flags
var (
	num_reqs    = flag.Int64("n", 10, "Number of requests.")
	concurrency = flag.Int("c", 1, "Concurrency for requests")
	url         = flag.String("u", "", "Url to send requests")
	verbose     = flag.Bool("v", false, "Show ongoing request results")
)

var (
	req_result        map[string]int64 = make(map[string]int64, 0)
	global_time       map[string]int64 = make(map[string]int64)
	response_times    []float64
	error_counts      int64
	current_job       int64
	status_codes      []int
	non_2xx           int64
	document_length   int64
	server_software   string
	is_first          bool
	total_transferred int64
	totalread         float64
)

func main() {
	flag.Parse()

	checkCommands()

	global_time["start"] = time.Now().UnixNano()

	current_job = 1
	for current_job <= *num_reqs {
		executeJobs(*concurrency)
	}

	global_time["end"] = time.Now().UnixNano()

	getStats()
}

func executeJobs(ccy int) {
	wg := new(sync.WaitGroup)
	for i := 1; i <= ccy; i++ {
		wg.Add(1)
		current_job += 1
		go sendRequest(*url, i, wg)
	}

	wg.Wait()
}

func sendRequest(url string, index int, wg *sync.WaitGroup) {
	defer wg.Done()

	start_time := time.Now().UnixNano()
	response, err := http.Get(url)
	if err != nil {
		if *verbose {
			fmt.Println(err)
		}
		error_counts += 1
	}

	if response.StatusCode > 206 {
		//addStatusCode(response.StatusCode)
		non_2xx += 1
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err == nil {
		totalread += float64(len(contents))
		total_transferred += int64(len(contents))

		if !is_first {
			is_first = true
			//fmt.Println(response.ContentLength)
			if len(response.Header["Server"]) > 0 {
				server_software = response.Header["Server"][0]
			}

			document_length = response.ContentLength
		}
		//totalread += float64(response.ContentLength)
	}

	//fmt.Println(response.ContentLength)
	response.Body.Close()

	end_time := time.Now().UnixNano()
	secs := toSecs(end_time - start_time)
	if err == nil {
		response_times = append(response_times, secs)
	}
	if *verbose {
		fmt.Println(fmt.Sprintf("Request sent: %.2f secs", secs))
	}
}

func getStats() {
	seconds := toSecs(global_time["end"] - global_time["start"])
	fmt.Println("Document Path:", *url)
	//fmt.Println("Document Length", document_length)
	fmt.Println("Server Software:", server_software)
	fmt.Printf("Time taken for tests: %.3f seconds\n", seconds)
	fmt.Printf("Completed requests: %d , Failed requests %.1f\n", current_job, percent(error_counts, current_job))
	fmt.Printf("Slowest reponse %.2f secs , Fastest response %.2f secs\n", findMax(), findMin())
	fmt.Println("Concurrency:", *concurrency)
	fmt.Printf("Requests per second %.2f\n", RequestPerSecond(seconds))
	fmt.Printf("Time per request %.2f\n", TimePerRequest(seconds))

	//fmt.Println(response_times)
	fmt.Println("Total Transfer:", total_transferred, "bytes")
	fmt.Printf("Transfer_rate: %.3f\n", TransferRate(seconds))
	if non_2xx > 0 {
		fmt.Println("Non-2xx:", non_2xx)
	}
}

func toSecs(secs int64) float64 {
	return float64(secs) / 1000000000.0
}

func percent(percent int64, total int64) float64 {
	return float64(percent) / float64(total) * 100.0
}

func findMax() float64 {
	var max float64 = 0

	for _, i := range response_times {
		if i > max {
			max = i
		}
	}

	return max
}

func findMin() float64 {
	var min float64 = 0

	for _, i := range response_times {
		if min == 0 {
			min = i
		}
		if i < min {
			min = i
		}
	}

	return min
}

func Usage() {
	fmt.Println("Hey there, you need to pass some options for this to work\n")
	fmt.Println("Flags:")
	flag.Parse()
	flag.PrintDefaults()
}

func checkCommands() {
	// No command? It's time for usage.
	if len(os.Args) == 1 {
		Usage()
		os.Exit(1)
	}

	flag.Parse()

	if *url == "" {
		fmt.Println("You need to pass an url to test, -u")
		Usage()
		os.Exit(1)
	}
}

func TransferRate(timetaken float64) float64 {
	return totalread / 1024 / timetaken
}

/*
Time per request
The average time spent per request.
*/
func TimePerRequest(seconds float64) float64 {
	return float64(*concurrency) * seconds / float64(current_job)
}

//Requests per second
/*This is the number of requests per second.
This value is the result of dividing the number of requests by the total time taken*/
func RequestPerSecond(seconds float64) float64 {
	var n float64
	n = float64(*num_reqs)
	return float64(n / float64(seconds))
}

// maybe convert this in map like {200 => 100 , 500 => 20, 401=> 3 }
func addStatusCode(status_code int) {
	exists := false
	for _, status := range status_codes {
		if status_code == status {
			exists = true
			break
		}
	}
	if !exists {
		status_codes = append(status_codes, status_code)
	}
}
