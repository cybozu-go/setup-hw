![CI](https://github.com/cybozu-go/setup-hw/workflows/main/badge.svg)

Hardware setup container
========================

This repository contains a Dockerfile and associated tools to build a
container image for configuring/monitoring server [BMC][] and [BIOS][].

Specifically, the image bundles `racadm` from [OMSA][] for Dell servers.

Usage
-----

### Build

```console
$ cd setup-hw
$ make build-image
```

### Run as a system service

The container need to be run as a system service before using `racadm` or [`setup-hw`](docs/setup-hw.md).

```console
$ sudo mkdir -p /var/lib/setup-hw

$ docker run -d --name=setup-hw \
  --net=host --privileged \
  -v /dev:/dev \
  -v /lib/modules:/lib/modules:ro \
  -v /etc/neco:/etc/neco:ro \
  -v /var/lib/setup-hw:/var/lib/setup-hw \
  setup-hw:latest
```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/setup-hw)

### Access `monitor-hw`

[`monitor-hw`](docs/monitor-hw.md) is the default command of the container.
When you run the container, it starts exporting hardware metrics for
Prometheus.  You can see the metrics from `http://localhost:9105/metrics`
by default.

You must prepare [configuration files](docs/config.md) before running
`monitor-hw`.

### Run racadm

```console
$ docker exec setup-hw racadm ...
```

### Run `setup-hw`

`setup-hw` is a tool to configure BMC and BIOS of the running server.
See the [document](docs/setup-hw.md) for details.


### Link
* [BMC](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller)
* [BIOS](https://en.wikipedia.org/wiki/BIOS)
* [OMSA](https://en.wikipedia.org/wiki/OpenManage#OMSA_%E2%80%93_OpenManage_Server_Administrator)
* [Prometheus](https://prometheus.io/)
* [Dell Remote Access Controller 9 RACADM CLI Guide](https://www.dell.com/support/manuals/ja-jp/poweredge-r7415/idrac9_7.xx_racadm_pub/introduction)
