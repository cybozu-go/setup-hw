Monitor hardware
================

`monitor-hw` is a monitoring daemon that exports [Redfish][] data for
Prometheus.


Synopsis
--------

```console
$ monitor-hw [--listen=<address>] [--interval=<interval>] [vendor-specific options...]
```


Description
-----------

`monitor-hw` is a monitoring daemon using [Redfish API][Redfish].
It works as an exporter for [Prometheus][] to monitor hardware metrics.

When `monitor-hw` is invoked, it starts a background routine to gather
Redfish data of the server where it runs.
This routine periodically traverses Redfish data from the BMC via HTTPS,
and stores them in memory.

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

`--listen=<address>` specifies the address, i.e. host and TCP port, where
`monitor-hw` listens for metrics retrieval.
This accepts the same forms of `<host>:<port>` and `:<port>` as `Addr` of
`http.Server`.
The default is `:9105`.

`--interval=<interval>` specifies the interval of Redfish data traversal
in seconds.
The default is 60 seconds.
This interval means the time between the end of a certain traversal operation
and the beginning of the next one, so the observed interval of metrics update
will be somewhat longer than this interval.

### Vendor specific options

#### Dell options

`--reset-interval` specifies the interval of resetting iDRAC in hours.
The default is 24 hours.


Configuration files
-------------------

`monitor-hw` reads the host and the user information of the BMC from
the configuration files.
See the [description of configuration files](config.md) for details.


Internals
---------

`monitor-hw` first detects the hardware type of the server where it runs.
It traverses and interprets Redfish data according to the collection rule
for that hardware type.
Collection rules are compiled from YAML files under
[redfish/rules](../redfish/rules).
See the [description of collection rules](rule.md) for details.


[Redfish]: https://www.dmtf.org/standards/redfish
[Prometheus]: https://prometheus.io/
