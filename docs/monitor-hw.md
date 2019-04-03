Monitor hardware
================

`monitor-hw` is a monitoring daemon which exports [Redfish][] data.


Synopsis
--------

```console
$ monitor-hw [--listen=<port>] [--redfish=<path>] [--interval=<interval>]
  [vendor-specific options...]
```


Description
-----------

`monitor-hw` is a monitoring daemon using [Redfish API][Redfish].
It works as an exporter for [Prometheus][] to monitor hardware metrics.

When `monitor-hw` is invoked, it starts a background routine to gather
Redfish data of the server where it runs.
This routine periodically traverses Redfish data which are published by
the BMC of the server via HTTPS, and stores them into its memory.

`monitor-hw` also starts an HTTP server to export data to Prometheus.
When its `/metrics` path is accessed, it reads the stored Redfish data
and reports them as the metrics of the server for Prometheus.

As `monitor-hw` uses the standard API of Redfish instead of dedicated tools
like `omreport`, it can support multiple types of servers from multiple
vendors.

### Vendor specific actions

#### Dell actions

`monitor-hw` starts `instsvcdrv-helper`.

`monitor-hw` periodically resets iDRAC because it occasinally hangs.


Options
-------

`--listen=<port>` specifies the TCP port number where `monitor-hw` listens
for metrics retrieval.  The default is 9105.

`--redfish=<path>` specifies the root path of the Redfish HTTPS URL.
The default is `/redfish/v1`.  Note that the host and the user information
are read from the configuration files.

`--interval=<interval>` specifies the interval of Redfish data traversal
in seconds.  The default is 60 seconds.

### Vendor specific options

#### Dell options

`--reset-interval` specifies the interval of resetting iDRAC in seconds.
The default is 3,600 seconds.


Configuration files
-------------------

`monitor-hw` reads the host and the user information of the BMC from
the configuration files.
See the [description of configuration files](config.md) for detail.


[Redfish]: https://www.dmtf.org/standards/redfish
[Prometheus]: https://prometheus.io/
