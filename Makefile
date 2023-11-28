build:
	go build -o build/ ./...

test:
	go test ./...

build-image: compile
	docker build -t jarvis .

clean:
	rm -rf build

compile:
	GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/ ./...

all: build
