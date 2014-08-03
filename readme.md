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
Process terminated: 3.631 seconds taken
Jobs processed: 11 , Failed 0.0
Max reponse 0.54 secs , Min response 0.33 secs
```

## Gotchas

In OSX you could find an limit connection issue, it could be fixed by doing

```sudo launchctl limit maxfiles 1000000 1000000```

To make this permanent (i.e not reset when you reboot), create /etc/launchd.conf containing:

In Unix/linux you can do:

```limit maxfiles 1000000 1000000```

##TODO

+ Bytes transferred
+ Configurable Timeouts
+ Document Length
+ Classify errors by failed reqs, broken pipes and exceptions
+ Response codes , show non 200 status
+ Transfer Rate
+ Change to a less boring name.

MIT license © Miguel Michelson 2014


