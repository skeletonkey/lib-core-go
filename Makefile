lint:
	go fmt ./...
	go vet ./...
	golangci-lint run --fix

deps-update:
	go get -u github.com/natefinch/lumberjack@latest
	go get -u github.com/rs/zerolog@latest

	go mod tidy

