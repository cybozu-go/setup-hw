[![Docker Repository on Quay](https://quay.io/repository/cybozu/setup-hw/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/setup-hw)
[![CircleCI](https://circleci.com/gh/cybozu-go/setup-hw.svg?style=svg)](https://circleci.com/gh/cybozu-go/setup-hw)

Hardware setup container
========================

This repository contains a Dockerfile and associated tools to build a
container image for configuring server [BMC][] and [BIOS][].

Specifically, the image bundles `idracadm7` from [OMSA][] for Dell servers.

Usage
-----

### Build

```console
$ docker build -t setup-hw:latest docker
```

### Run as a system service

The container need to be run as a system service before using `idracadm7` or `setup-hw`.

rkt and systemd:

```console
$ sudo systemd-run --unit=setup-hw.service \
  rkt run --net=host --dns=host --hosts-entry=host --hostname=%H \
  --insecure-options=all \
  --volume dev,kind=host,source=/dev --mount volume=dev,target=/dev \
  --volume sys,kind=host,source=/sys --mount volume=sys,target=/sys \
  --volume modules,kind=host,source=/lib/modules,readOnly=true --mount volume=modules,target=/lib/modules \
  --volume neco,kind=host,source=/etc/neco,readOnly=true --mount volume=neco,target=/etc/neco \
  setup-hw:latest \
    --name setup-hw \
    --caps-retain=CAP_SYS_ADMIN,CAP_SYS_CHROOT,CAP_CHOWN,CAP_FOWNER,CAP_NET_ADMIN
```

Docker:

```console
$ docker run -d --name=setup-hw \
  --net=host --privileged \
  -v /dev:/dev \
  -v /lib/modules:/lib/modules:ro \
  -v /etc/neco:/etc/neco:ro \
  setup-hw:latest
```

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

Hardware auto configuration
---------------------------

The container image includes a tool `setup-hw` to configure BMC and BIOS of the running server.
`setup-hw` reads following files:

### `/etc/neco/bmc-address.json`

The contents is a JSON object like this:

```json
{
    "ipv4": {
        "address": "1.2.3.4",
        "netmask": "255.255.255.0",
        "gateway": "1.2.3.1"
    }
}
```

BMC network interface will be configured to have the given `address`.

### `/etc/neco/bmc-user.json`

This file contains credentials of BMC users.

BMC users are statically defined in `setup-hw` as follows:

* `root`: The administrator of BMC.
* `power`: Control power supply.
* `support`: Read-only account.

Credential types are:

* Raw password
* Hashed password with salt  
    For iDRAC, use [`idrac-passwd-hash`](./pkg/idrac-passwd-hash) tool to generate them.
* Authorized public keys for SSH

Supported credential types varies by BMC types.
iDRAC, BMC embedded in Dell servers, supports all credential types.

Example:

```json
{
    "root": {
        "password": {
            "raw": "raw password"
        },
        "authorized_keys": [
            "ssh-rsa ...",
            ...
        ]
    },
    "power": {
        "password": {
            "hash": "hashed_secret",
            "salt": "salt for hash"
        }
    }
}
```

### How to run `setup-hw`

1. Run `setup-hw` container as a system service.
2. Prepare `/etc/neco/bmc-address.json` and `/etc/neco/bmc-user.json`.
3. Use `rkt enter` or `docker exec` to run `setup-hw` inside the container.
4. If `setup-hw` exits with status code 10, the server need to be rebooted.

rkt:

```console
$ sudo rkt enter $POD_UUID setup-hw
$ if [ $? -eq 10 ]; then sudo reboot; done
```

Docker:

```console
$ docker exec setup-hw setup-hw
$ if [ $? -eq 10 ]; then sudo reboot; done
```

[BMC]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[BIOS]: https://en.wikipedia.org/wiki/BIOS
[OMSA]: https://en.wikipedia.org/wiki/OpenManage#OMSA_%E2%80%93_OpenManage_Server_Administrator
