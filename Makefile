build:
	go build -o bin/goplay ./...

setup: 
	go get github.com/stretchr/testify/assert

fmt:
	gofmt -s -w .

test:
	go test -v -race -coverpkg=./... -coverprofile=coverage.txt -covermode=atomic ./...

run:	build
	bin/goplay

clean:
	rm -rf bin coverage.txt
