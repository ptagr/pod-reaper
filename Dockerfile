FROM golang:1.13.4-alpine

RUN apk --no-cache add curl git && curl https://glide.sh/get | sh && apk del curl

WORKDIR /go/src/app
COPY ["glide.yaml", "main.go", "/go/src/app/"]

RUN glide install

RUN go build -o /pod-reaper

CMD  /pod-reaper
