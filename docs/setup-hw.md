Hardware Auto Configuration
===========================

`setup-hw` is a tool to configure BMC and BIOS of the running server.


How to run `setup-hw`
---------------------

1. Run `setup-hw` container as a system service.  See [README](../README.md).
2. Prepare `/etc/neco/bmc-address.json` and `/etc/neco/bmc-user.json`.  See [config page](config.md).
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
