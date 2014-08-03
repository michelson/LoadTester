#Load Tester:

This is a simple library to test web load, right now it only supports get requests but more verbs, cookie support and basic auth are on its way.


##Usage:

Build library with ```go build main.go```

```
Flags:
  -c=1: Concurrency for requests
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

In OSX you could find a limit connection issue, it could be fixed by doing

```sudo launchctl limit maxfiles 1000000 1000000```

To make this permanent (i.e not reset when you reboot), create /etc/launchd.conf containing:

```limit maxfiles 1000000 1000000```

##TODO

+ Bytes transferred
+ Configurable Timeouts
+ Document Length
+ Classify errors by failed reqs, broken pipes and exceptions
+ Transfer Rate
+ Change to a less boring name.

MIT license © Miguel Michelson 2014


