# Prometheus scrape endpoint generator

Application to be run as part of a Prometheus load test which exposes multiple target endpoints for Prometheus to scrape.

Endpoints are provided at `/metrics/<endpoint_number>` for example `/metrics/2`. The metrics returned include a range of
counters, gauges and histograms with random values.

This tool is designed to be used in conjunction with the load test tool and the [Prometheus terraform repo](https://github.com/alphagov/prometheus-aws-configuration-beta).

It is based off the [Prometheus prombench fake-webserver](https://github.com/prometheus/prombench/tree/master/components/prombench/apps/fake-webserver) and has been adapted to make it easier to generate the necessary metrics and size of time series needed for load testing at a number of different loads.

## Configuration

- `ENDPOINT_COUNT` - The number of endpoints to expose from `/metrics/1` to `/metrics/<ENDPOINT_COUNT>`
- `ROUTE_COUNT` - The number of routes for our pretend application on which the metrics are based (roughly 1 route equates to 60 metrics)
- `SCRAPE_INTERVAL` - The scrape interval in seconds for Prometheus of your targets

## Pre-requisites

You will need to install [Go lang](https://golang.org/doc/install)

## Running the webserver
Make a copy of `environment_sample.sh` to `environment.sh` file and source it - `source environment.sh`.

Execute `./run_app.sh -w` on a terminal to run the webserver locally.

## Deploying the server to PaaS

Update the `manifest.yml` to set your desired `ENDPOINT_COUNT` and `ROUTE_COUNT`.

Run `cf push` in the `sandbox` space.

## Set up your Prometheus to scrape your load test endpoints

Execute `./run_app.sh -c` on a terminal to create the scrape config section for your new targets. It will look similar to:

```
- job_name: load-test-targets-1
  metrics_path: "/metrics/1"
  scheme: https
  scrape_interval: 30s
  static_configs:
  - targets: ["prom-loadtest-metrics-generator.cloudapps.digital"]
- job_name: load-test-targets-2
  metrics_path: "/metrics/2"
  scheme: https
  scrape_interval: 30s
  static_configs:
  - targets: ["prom-loadtest-metrics-generator.cloudapps.digital"]
```

You should then append this to the config file for your development Prometheus stack at `terraform/app-ecs-services/templates/prometheus.tpl`.
