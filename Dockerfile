FROM golang:1.24 AS builder
ARG TARGET
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o build/app $TARGET


FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/build/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
