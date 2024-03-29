.DEFAULT_GOAL = build

build:
	@echo "Building the binary..."
	mkdir -p bin
	go build -o bin/skywoker cmd/main/main.go

run:
	@echo "Running skywoker binary..."
	go run bin/skywoker

compile:
	@echo "Compiling for every OS and Platform"
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/skywoker-linux-amd64 cmd/main/main.go
	GOOS=linux GOARCH=arm go build -o bin/skywoker-linux-arm cmd/main/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/skywoker-linux-arm64 cmd/main/main.go
	GOOS=freebsd GOARCH=386 go build -o bin/skywoker-freebsd-386 cmd/main/main.go

all: build
