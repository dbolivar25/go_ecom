build:
	@go build -C src/ -o ../bin/go_ecom

run: build
	@./bin/go_ecom

test:
	@go test -v ./...
