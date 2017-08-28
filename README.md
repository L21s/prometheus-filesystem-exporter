# Prometheus Filesystem Exporter
A very simple exporter that reads files from a given directory and exports their contents as untyped metrics.

## Usage

The easiest way is to use docker:

```shell
docker run \ 
-v /host/metrics:/container/metrics \
-p 8080:8080 \
larscheidschmitzhermes/prometheus-filesystem-exporter:latest \
--metrics-directory /container/metrics \
--metrics-path /metrics \
--listen-addr :8080
```

To add a new metric, simply write a file with it's name (_optionally_ followed by the labels you want to assign) to the `metrics` directory:

```shell
touch '/host/metrics/answer_to_everything;scope=universe;env=prod'
```

You will see that the exporter immediately picks up the metric:

```shell
curl -s http://localhost:8080/metrics | grep answer_to_everything
# HELP answer_to_everything Auto generated from filesystem path: /container/metrics/answer_to_everything
# TYPE answer_to_everything untyped
answer_to_everything{scope="universe",env="prod"} 0
```

Setting the value is as easy as writing to the file:

```shell
echo "42" > '/host/metrics/answer_to_everything;scope=universe;env=prod'
```

The update happens right away:

```shell
curl -s http://localhost:8080/metrics | grep answer_to_everything
# HELP answer_to_everything Auto generated from filesystem path: /container/metrics/answer_to_everything
# TYPE answer_to_everything gauge
answer_to_everything{scope="universe",env="prod"} 42
```

## Configuration

The following flags can be passed to the `prometheus-filesystem-exporter`:

| flag                  | default   | description   |
| ---                   | ---       | ---           |
| `--listen-addr`         | `:8080`     | The address to listen on for http requests
| `--metrics-directory`   | `/metrics`  | The directory to read metrics from |
| `--metrics-path`        | `/metrics`  | The http path under which metrics are exposed