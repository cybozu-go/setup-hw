Automated Firmware Update Support Tool
======================================

`setup-apply-firmware` is a tool to configure BMC to apply firmware update.

Synopsis
--------

```console
$ setup-apply-firmware UPDATER_URL...
```

`UPDATER_URL` should be those supported by `curl`.

Description
-----------

`setup-apply-firmware` is a tool to configure BMC to apply firmware update.

It downloads the firmware updaters specified by `UPDATER_URL`s and sends them to BMC.

Caveat
------

- The order to send each updater is indefinite.
- This tool does not initiate reboot.
