BINS := sniffer-bitcoin sniffer-ethereum

build:
	go build -o build/ ./...

test:
	go test ./...


dockerbuild-%:
	docker build --build-arg TARGET=./cmd/$* -t $* .

docker: $(addprefix dockerbuild-,$(BINS))


clean:
	rm -rf build

