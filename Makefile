default: run

build: fmt
	go build -o skydo *.go

run: build
	./skydo

fmt:
	gofmt -s -l .
