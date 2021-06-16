FROM golang:latest AS builder

WORKDIR /go/src/app
COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o pod-reaper

FROM alpine:latest

COPY --from=builder /go/src/app/pod-reaper /pod-reaper
CMD /pod-reaper
