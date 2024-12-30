.PHONY: build test milk

build:
	go fmt . &&  go build

milk: *.go pm/*.go cmd/milk/milk.go
	go fmt . && go build cmd/milk/milk.go

test:
	go fmt . &&  go test
