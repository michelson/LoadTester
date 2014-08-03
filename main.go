package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var num_reqs = flag.Int64("n", 10, "Number of requests.")
var concurrency = flag.Int("c", 1, "Concurrency for requests")
var url = flag.String("u", "", "Url to send requests")
var verbose = flag.Bool("v", false, "Show ongoing request results")

var req_result map[string]int64 = make(map[string]int64, 0)
var global_time map[string]int64 = make(map[string]int64)
var response_times []float64
var error_counts int64
var current_job int64

func main() {

	checkCommands()

	flag.Parse()

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
	_, err := http.Get(url)
	if err != nil {
		if *verbose {
			fmt.Println(err)
		}
		error_counts += 1
	}
	end_time := time.Now().UnixNano()
	secs := toSecs(end_time - start_time)
	response_times = append(response_times, secs)
	if *verbose {
		fmt.Println(fmt.Sprintf("sent job %.2f", secs))
	}
}

func getStats() {
	seconds := toSecs(global_time["end"] - global_time["start"])

	fmt.Println(fmt.Sprintf("Process terminated: %.3f seconds taken", seconds))
	fmt.Println(fmt.Sprintf("Jobs processed: %d , Failed %.1f", current_job, percent(error_counts, current_job)))
	fmt.Println(fmt.Sprintf("Max reponse %.2f secs , Min response %.2f secs", findMax(), findMin()))
	//fmt.Println(response_times)
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
	fmt.Println("Hey there , you need to pass some options for this to work")
	fmt.Println("\nFlags:")
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
