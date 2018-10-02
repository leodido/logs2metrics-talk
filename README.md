# Talk: From logs to metric

This repository contains the PoC associated with the talk "From logs to metric".

Its main goal is to show how to extract (structured) value from the huge amount of (unstructured) information that logs contain.

In brief, the steps are as follows: parsing of syslog messages into structured data, ingesting/collecting them via Telegraf syslog input plugin, visualizing and plot them via Chronograf's log viewer, and eliciting new meaningful metrics to plot processing them via Kapacitor.

The stack used to achieve this is:

- [Telegraf](https://github.com/influxdata/telegraf) with [syslog input plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/syslog), which uses this blazing fast [go-syslog](https://github.com/influxdata/go-syslog) parser
- Chronograf
- InfluxDB
- [Kapacitor](https://github.com/influxdata/kapacitor)

## Setup

First of all we need a local k8s environment.

Let's proceed with minikube.

```bash
minikube start --docker-opt log-driver=journald
```

Note that we need the journald log driver for the inner docker since the rsyslog's mmkubernetes module [only works with it](https://www.rsyslog.com/doc/master/configuration/modules/mmkubernetes.html) (or with json-file docker log driver).

The following step is to become a YAML developer :hear_no_evil: :speak_no_evil:, applying all the YAML files describing our setup.

```bash
kubectl apply -f ...
```

Finally to access Chronograf from within our local browser we need the following port forward.

```bash
kubectl port-forward svc/chronograf -n logging 8888:80
```

Go to [localhost:8888](http://localhost:8888) now!

## Developing the Kapacitor UDF

File `docker-compose.yaml` is useful during the development and debugging of the Kapacitor UDF.

To make it working do not forget to forward the port of the influxdb within minikube.

```bash
kubectl port-forward svc/influxdb -n logging 8686:8686
```

Then run

```bash
docker-compose up -d
```

---

[![Analytics](https://ga-beacon.appspot.com/UA-49657176-1/logs2metrics-talk?flat)](https://github.com/igrigorik/ga-beacon)