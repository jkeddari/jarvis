FROM golang:1.21 as dev

WORKDIR /jarvis
COPY . /jarvis

RUN make clean && make build


FROM debian:stable
COPY --from=dev /jarvis/build/server /jarvis/
ENTRYPOINT ["/jarvis/server"]

