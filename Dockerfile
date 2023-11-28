FROM golang:1.21 as base

FROM base as dev


ADD build/linux-arm64/server /jarvis/server
WORKDIR /jarvis

CMD ["./server"]
