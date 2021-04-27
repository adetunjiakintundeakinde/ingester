# Introduction
This is basically my answer described to the problem in Screening_Task.pdf. It's not a production grade code as it was done in limited time.

# Ingester

Basically starts the dispatcher and the handlers

# Getting started

To start a server listening at port **9020** and requiring basic authentication 
run the following set of commands:

```console
$ make build
$ ./ingester
```

You can the metrics and view the state of the internal components such as current values and errors

```console
$ curl localhost:9020/metrics | jq .
```

Data can look as follows

```json
{
  "cpu_usage_average_cpu": {
    "value": 0.49989037293819977
  },
  "cpu_usage_errors": {
    "count": 0
  },
  "cpu_usage_payload_processed_count": {
    "count": 55946
  },
  "dispatcher_errors": {
    "count": 0
  },
  "dispatcher_payload_processed_count": {
    "count": 168986
  },
  "last_kernel_upgrade_errors": {
    "count": 0
  },
  "last_kernel_upgrade_latest_time": {
    "value": 1613860787
  },
  "last_kernel_upgrade_payload_processed_count": {
    "count": 56357
  },
  "load_avg_errors": {
    "count": 0
  },
  "load_avg_maximum_load": {
    "value": 0.999986
  },
  "load_avg_minimum_load": {
    "value": 0.000020052788
  },
  "load_avg_payload_processed_count": {
    "count": 56683
  }
}
```
# Architecture
Dispatchers and handlers run on different go routines but once the process is canceled in the main process the dispatchers and handlers are stopped

# Dispatcher
Dispatcher calls the ingester and errors increase if it cannot marshal the data or receives server errors from the ingesters

# Handlers
There are three handlers:
1. Cpu 
2. Last Kernel Upgrade
3. Load Average

Handlers increase errors when the data in the payload is not what it expects