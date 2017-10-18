# pod-reaper
A kubernetes operator that reaps pods that have reached their lifetime

# Configuration
To give a lifetime to your pods, add the following annotation:

`pod.kubernetes.io/lifetime: $DURATION`

`DURATION` has to be a [valid golang duration string](https://golang.org/pkg/time/#ParseDuration).

A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Example: `kubernetes.io/lifetime: 720h` 

The above annotation will cause the pod to be reaped (killed) once it reaches the age of 30d (720h)