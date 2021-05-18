Automated ISO Image Reboot Support Tool
======================================

`setup-isoreboot` is a tool to configure BMC to reboot from a ISO image.

Synopsis
--------

```console
$ setup-isoreboot ISO_IMAGE_URL
```

Description
-----------

`setup-isoreboot` is a tool to configure BMC to reboot from a ISO image.

It connects the ISO image to Virtual CD/DVD and make it next boot device once.

Caveat
------

- This tool does not initiate reboot.
