# Talk: From logs to metric with the TICK stack

[**Slides**](http://bit.ly/from-logs-to-metrics-tick).

This repository contains the PoC associated with the talk "From logs to metric with the TICK stack".

Its main goal is to show how to extract (structured) value from the huge amount of (unstructured) information that logs contain.

In brief, the steps are as follows: parsing of syslog messages into structured data, ingesting/collecting them via Telegraf syslog input plugin, visualizing and plot them via Chronograf's log viewer, and eliciting new meaningful metrics (eg. number of process OOM killed) to plot processing them via a Kapacitor [UDF](https://docs.influxdata.com/kapacitor/v1.5/guides/socket_udf/).

The stack used to achieve this is:

- [Telegraf](https://github.com/influxdata/telegraf) with [syslog input plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/syslog), which uses this blazing fast [go-syslog](https://github.com/influxdata/go-syslog) parser
- Chronograf
- InfluxDB
- [Kapacitor](https://github.com/influxdata/kapacitor)

![Chronograf Log Viewer](images/logviewer-chronograf.png "Chronograf Log Viewer")

![Exploring RFC5425 syslog messages with Chronograf](images/exploring-syslog-chronograf.png "Exploring RFC5425 syslog messages with Chronograf")

![Couting OOMs](images/ooms-num.png "Couting OOMs")

![Counting OOMs of stress pod](images/ooms-stress.png "Counting OOMs of stress pod")


## Setup

First of all we need a local k8s environment.

Let's proceed with minikube.

```bash
minikube start --docker-opt log-driver=journald
```

Note that we need the journald log driver for the inner docker since the rsyslog's mmkubernetes module [only works with it](https://www.rsyslog.com/doc/master/configuration/modules/mmkubernetes.html) (or with json-file docker log driver).

The following step is to become a YAML developer :hear_no_evil: :speak_no_evil:, applying all the YAML files describing our setup.

<div style="display: flex; align-items: center;">
<div style="flex: 33.33%; padding: 4px;">
<img alt="YAML meme" src="images/yaml-dev.jpg">
</div>
<div style="flex: 33.33%; padding: 4px;">
<img alt="The life of a YAML developer" src="images/yaml-dev-life.jpg">
</div>
</div>


And execute the following commands.

```bash
kubectl apply -f namespace.yaml
kubectl apply -f roles.yaml
kubectl apply -f influxdb.yaml
kubectl apply -f telelog.yaml
kubectl apply -f chronograf.yaml
kubectl apply -f kapacitor.yaml
kubectl apply -f stress.yaml
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