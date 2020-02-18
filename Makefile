default: run

build:
	go build -o skydo *.go

run: build
	./skydo
