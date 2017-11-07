FROM golang:alpine

COPY pod-reaper /pod-reaper
CMD  /pod-reaper