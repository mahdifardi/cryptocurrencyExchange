build:
	go build -o ./bin/exchange

run:
	go run main.go

test:
	go test -v ./...