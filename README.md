[![Docker Repository on Quay](https://quay.io/repository/cybozu/setup-hw/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/setup-hw)

Hardware setup container
========================

This repository contains a Dockerfile and associated tools to build a
container image for configuring server [BMC][] and [BIOS][].

Specifically, the image bundles `idracadm7` from [OMSA][] for Dell servers.

Usage
-----

### Build

```console
$ docker build -t setup-hw:latest .
```

### Run as a system service

rkt:

```console
$ sudo rkt run --net=host --dns=host --hosts-entry=host --hostname=%H \
  --insecure-options=all \
  --volume cg,kind=host,source=/sys/fs/cgroup --mount volume=cg,target=/sys/fs/cgroup \
  --volume neco,kind=host,source=/etc/neco,readOnly=true --mount volume=neco,target=/etc/neco \
  setup-hw:latest \
    --name setup-hw
```

Docker:

```console
$ docker run -d --name=setup-hw \
  --net=host --privileged \
  -v /sys/fs/cgroup:/sys/fs/cgroup \
  -v /etc/neco:/etc/neco:ro \
  setup-hw:latest
```

### Run idracadm7

rkt:

```console
$ sudo rkt enter POD_UUID idracadm7 ...
```

Docker:

```console
$ docker exec setup-hw idracadm7 ...
```



[BMC]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[BIOS]: https://en.wikipedia.org/wiki/BIOS
[OMSA]: https://en.wikipedia.org/wiki/OpenManage#OMSA_%E2%80%93_OpenManage_Server_Administrator
