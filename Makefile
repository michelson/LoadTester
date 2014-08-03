default:	build

build:
	go build main.go

run:
	go run main.go -c=1 -n=10 -u=http://golang.org -v=true

test:
	go test

format:
	go fmt