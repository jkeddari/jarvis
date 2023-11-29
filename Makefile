build:
	go build -o build/ ./...

test:
	go test ./...

build-image: compile
	docker build -t jarvis .

clean:
	rm -rf build

all: build
