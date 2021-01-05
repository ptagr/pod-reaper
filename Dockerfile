FROM golang:1.15.6-alpine

COPY pod-reaper /pod-reaper
CMD  /pod-reaper