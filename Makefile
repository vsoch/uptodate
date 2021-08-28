all:
	gofmt -s -w .
	go build -o uptodate
	
build:
	go build -o uptodate
	
run:
	go run main.go
