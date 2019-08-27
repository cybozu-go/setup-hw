[![Docker Repository on Quay](https://quay.io/repository/cybozu/setup-hw/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/setup-hw)
[![CircleCI](https://circleci.com/gh/cybozu-go/setup-hw.svg?style=svg)](https://circleci.com/gh/cybozu-go/setup-hw)

Hardware setup container
========================

This repository contains a Dockerfile and associated tools to build a
container image for configuring/monitoring server [BMC][] and [BIOS][].

Specifically, the image bundles `idracadm7` from [OMSA][] for Dell servers.

Usage
-----

### Build

```console
$ docker build -t setup-hw:latest docker
```

### Run as a system service

The container need to be run as a system service before using `idracadm7` or [`setup-hw`](docs/setup-hw.md).

rkt and systemd:

```console
$ sudo mkdir -p /var/lib/setup-hw

$ sudo systemd-run --unit=setup-hw.service \
  rkt run --net=host --dns=host --hosts-entry=host --hostname=%H \
  --insecure-options=all \
  --volume dev,kind=host,source=/dev --mount volume=dev,target=/dev \
  --volume sys,kind=host,source=/sys --mount volume=sys,target=/sys \
  --volume modules,kind=host,source=/lib/modules,readOnly=true --mount volume=modules,target=/lib/modules \
  --volume neco,kind=host,source=/etc/neco,readOnly=true --mount volume=neco,target=/etc/neco \
  --volume var,kind=host,source=/var/lib/setup-hw --mount volume=var,target=/var/lib/setup-hw \
  setup-hw:latest \
    --name setup-hw \
    --caps-retain=CAP_SYS_ADMIN,CAP_SYS_CHROOT,CAP_CHOWN,CAP_FOWNER,CAP_NET_ADMIN
```

Docker:

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

### Access `monitor-hw`

[`monitor-hw`](docs/monitor-hw.md) is the default command of the container.
When you run the container, it starts exporting hardware metrics for
Prometheus.  You can see the metrics from `http://localhost:9105/metrics`
by default.

You must prepare [configuration files](docs/config.md) before running
`monitor-hw`.

### Run idracadm7

rkt:

```console
$ POD_UUID=$(sudo rkt list --full | grep running | grep setup-hw | cut -f 1)
$ sudo rkt enter $POD_UUID idracadm7 ...
```

Docker:

```console
$ docker exec setup-hw idracadm7 ...
```

### Run `setup-hw`

`setup-hw` is a tool to configure BMC and BIOS of the running server.
See the [document](docs/setup-hw.md) for details.


[BMC]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[BIOS]: https://en.wikipedia.org/wiki/BIOS
[OMSA]: https://en.wikipedia.org/wiki/OpenManage#OMSA_%E2%80%93_OpenManage_Server_Administrator
[Prometheus]: https://prometheus.io/
