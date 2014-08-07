#Load Tester:

This is a simple package to benchmark http requests,  This especially shows you how many requests per second a http server is capable of serving.

##Usage:

Build library with ```go build main.go```

```
Flags:
  -A="": Supply BASIC Authentication credentials to the server. The username and password are separated by a single : and sent on the wire base64 encoded. The string is sent regardless of whether the server needs it (i.e., has sent an 401 authentication needed).
  -F="": Add a Cookie from plain text. The file should contain multiple cookie information separated by line
  -H="": Custom headers name:value;name2:value2
  -T="text/html": Content type, default to text/html
  -c=1: Number of multiple requests to perform at a time. Default is one request at a time.
  -n=10: Number of requests.
  -u="": Url to send requests
  -v=false: Show ongoing request results
```

###Example:

```main -c=1 -n=10 -u=http://golang.org -v=true```

##Response:
```
Document Path: http://domain.com
Server Software: nginx/1.4.6 (Ubuntu)
Time taken for tests: 1.505 seconds
Completed requests: 9 , Failed requests 0.0 %
Slowest reponse 0.59 secs , Fastest response 0.30 secs
Concurrency: 2
Requests per second 5.32
Time per request 0.33
Total Transfer: 50544 bytes
Transfer_rate: 32.804
```

## Gotchas

In OSX you may encounter an limit connection issue, it could be fixed by doing

```sudo launchctl limit maxfiles 1000000 1000000```

To make this permanent (i.e not reset when you reboot), create /etc/launchd.conf containing:

```limit maxfiles 1000000 1000000```

##TODO

+ Bytes transferred
+ Configurable Timeouts
+ Document Length
+ Classify errors by failed reqs, broken pipes and exceptions
+ Report status codes
+ Export to CSV
+ Transfer Rate
+ Change to a less boring name.

MIT license © Miguel Michelson 2014


