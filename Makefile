build:
	@go build -o bin/fiabesco-backend cmd/main.go

run: build
	@go run cmd/main.go