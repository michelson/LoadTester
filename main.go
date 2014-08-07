package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"github.com/rakyll/pb"
)

// Flags
var (
	num_reqs    = flag.Int64("n", 10, "Number of requests.")
	concurrency = flag.Int("c", 1, "Number of multiple requests to perform at a time. Default is one request at a time.")
	url         = flag.String("u", "", "Url to send requests")
	verbose     = flag.Bool("v", false, "Show ongoing request results")
	header_line = flag.String("H", "", "Custom headers name:value;name2:value2")
	//cookie      = flag.String("C", "", "Add a Cookie: line to the request. The argument is typically in the form of a name=value pair. This field is repeatable.")
	cookie_file  = flag.String("F", "", "Add a Cookie from plain text. The file should contain multiple cookie information separated by line")
	auth         = flag.String("A", "", "Supply BASIC Authentication credentials to the server. The username and password are separated by a single : and sent on the wire base64 encoded. The string is sent regardless of whether the server needs it (i.e., has sent an 401 authentication needed).")
	content_type = flag.String("T", "text/html", "Content type, default to text/html")
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

var client = &http.Client{}
var (
	cookies  = make([]*http.Cookie, 0)
	username string
	password string
	headers  [][]string
)

var Bar *pb.ProgressBar

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	Bar = pb.New(int(*num_reqs))
	Bar.Format("ooo")
	//Bar.Format("")

	//Bar.SetUnits(pb.U_BYTES)

	checkCommands()

	checkReqOptions()

	global_time["start"] = time.Now().UnixNano()

	current_job = 1

	Bar.Start()

	for current_job <= *num_reqs {
		executeJobs(*concurrency)
	}

	global_time["end"] = time.Now().UnixNano()
	Bar.FinishPrint("Tasks completed")

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
	Bar.Increment()

	start_time := time.Now().UnixNano()
	//response, err := http.Get(url)
	req := Request(url)
	response, err := client.Do(req)
	if err != nil {
		if *verbose {
			fmt.Println(err)
		}
		error_counts += 1
	}
	if err != nil {
		return
	}
	if response.StatusCode > 206 {
		addStatusCode(response.StatusCode)
		non_2xx += 1
	}

	contents, err := ioutil.ReadAll(response.Body)
	//fmt.Println(string(contents))
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
		//fmt.Println(fmt.Sprintf("Request sent: %.2f secs", secs))
	}
}

func getStats() {
	seconds := toSecs(global_time["end"] - global_time["start"])
	fmt.Println("Document Path:", *url)
	//fmt.Println("Document Length", document_length)
	fmt.Println("Server Software:", server_software)
	fmt.Printf("Time taken for tests: %.3f seconds\n", seconds)
	fmt.Printf("Completed requests: %d, Failed requests: %.1f\n", current_job, percent(error_counts, current_job))
	fmt.Printf("Slowest reponse %.2f secs, Fastest response: %.2f secs\n", findMax(), findMin())
	fmt.Println("Concurrency:", *concurrency)
	fmt.Printf("Requests per second %.2f\n", RequestPerSecond(seconds))
	fmt.Printf("Time per request %.2f\n", TimePerRequest(seconds))

	//fmt.Println(response_times)
	//fmt.Println("status_codes", status_codes)
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

func checkReqOptions() {
	if *cookie_file != "" {
		cookies = parseCookieFile()
	}

	if *header_line != "" {
		headers = parseHeaders()
	}

	if *auth != "" {
		username, password = parseBasicAuth()
	}
}

//parse cookie file
func parseCookieFile() []*http.Cookie {
	lines := ReadLines(*cookie_file)
	cookies := make([]*http.Cookie, 0)
	for _, line := range lines {
		line_res := make(map[string]string)
		options := strings.Split(line, ";")
		for _, option := range options {
			key_value := strings.Split(option, "=")
			key_value[0] = strings.Trim(key_value[0], " ")
			switch key_value[0] {
			case "path":
				line_res["path"] = key_value[1]
			case "domain":
				line_res["domain"] = key_value[1]
			//case "expires":
			//    line_res["expires"] = key_value[1]
			default:
				line_res["name"] = key_value[0]
				line_res["value"] = key_value[1]
			}
		}
		c := &http.Cookie{}
		c.Name = line_res["name"]
		c.Value = line_res["value"]
		c.Domain = line_res["domain"]
		c.Path = line_res["path"]
		cookies = append(cookies, c)
	}
	//fmt.Println("cookies", cookies)
	return cookies
}

//parse auth
func parseBasicAuth() (string, string) {
	auth := strings.Split(*auth, ":")
	if len(auth) > 2 {
		return auth[0], auth[1]
	} else {
		return "", ""
	}
}

//parse headers
func parseHeaders() [][]string {
	options := strings.Split(*header_line, ";")
	response := make([][]string, 0)
	for _, option := range options {
		o := strings.Split(option, ":")
		response = append(response, o)
	}
	return response
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

// REQUEST
func Request(url string) *http.Request {

	method := "GET"
	req, _ := http.NewRequest(method, url, nil)

	//update the Host value in the Request - this is used as the host header in any subsequent request
	//req.Host = r.OriginalHost
	req.Header.Add("Content-Type", *content_type)

	for _, c := range cookies {
		req.AddCookie(c)
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	for _, h := range headers {
		req.Header.Add(h[0], h[1])
	}

	//resp, err := client.Do(req)
	return req
}

func ReadLines(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	var response []string
	bf := bufio.NewReader(f)

	for {
		line, isPrefix, err := bf.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		if isPrefix {
			log.Fatal("Error: Unexpected long line reading", f.Name())
		}

		//fmt.Println(string(line))
		response = append(response, string(line))
	}
	return response
}
