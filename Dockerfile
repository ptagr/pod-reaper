FROM alpine

COPY pod-reaper /pod-reaper
CMD  /pod-reaper