# stage: build ---------------------------------------------------------

FROM golang:1.22-alpine as build

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/node-monitor

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -o bin/node-monitor -ldflags "-s -w" github.com/flashbots/node-monitor/cmd

# stage: run -----------------------------------------------------------

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/node-monitor/bin/node-monitor ./node-monitor

ENTRYPOINT ["/app/node-monitor"]
