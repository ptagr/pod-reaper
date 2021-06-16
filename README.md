# pod-reaper

A kubernetes operator that reaps

* pods that have reached their lifetime
* evicted pods

## Configuration

To give a lifetime to your pods, add the following annotation:

`pod.kubernetes.io/lifetime: $DURATION`

`DURATION` has to be a [valid golang duration string](https://golang.org/pkg/time/#ParseDuration).

A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Example: `kubernetes.io/lifetime: 720h`

The above annotation will cause the pod to be reaped (killed) once it reaches the age of 30d (720h)

## Deployment Details

The pod reaper needs to be deployed in a kubernetes cluster.

The following environment variables can be set:

| Env Variable           | Description                                                                                  | Sample Values         | Default value | Required |
|------------------------|----------------------------------------------------------------------------------------------|-----------------------|---------------|----------|
| REMOTE_EXEC            | Should be set to true when running within the cluster, to false when running locally         | true                  | N/A           | yes      |
| REAPER_NAMESPACES      | List of namespaces that the reaper would inspect                                             | namespace1,namespace2 | N/A           | yes      |
| CRON_JOB               | Whether this should be run just once or in a loop. Set to true if running this as a cron job | true                  | false         | no       |
| MAX_REAP_COUNT_PER_RUN | Maximum Pods to reap in each run                                                             | 100                   | 30            | no       |
| REAP_EVICTED_PODS      | Whether or not to delete evicted pods                                                        | true                  | false         | no       |
| EVICT                  | Use Eviction instead of Deletion when removing pods (honors Pod Disruption Budgets)          | true                  | false         | no       |

## Todo

* Support RBAC
