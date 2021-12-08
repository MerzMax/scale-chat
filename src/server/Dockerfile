FROM golang:buster AS builder
WORKDIR /app

COPY ./src/go.mod ./src/go.sum ./
RUN go mod download

COPY ./src .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/out/server ./server

FROM alpine:3.14
WORKDIR /app

RUN adduser -S app
USER app

COPY --from=builder /app/out/server ./
COPY --from=builder /app/server/demo.html ./demo.html

EXPOSE 8080
ENTRYPOINT ["./server"]