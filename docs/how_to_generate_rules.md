How to Generate Data Collection Rules
=====================================

[monitor-hw](monitor-hw.md) exports metrics for [Prometheus][] by accessing
[Redfish][] API and converting retrieved data according to [data collection
rules](rule.md).
It uses YAML files under [redfish/rules](../redfish/rules) as sources of rules.

You can write a new rule file for a new type of servers by yourself, or you
can use [collector](collector.md) command-line tool.
This document describes how to use `collector` to generate a rule file.


Update Server Firmware
----------------------

First of all, make sure all servers have up-to-date BIOS and BMC.
Follow the instructions below to update them if you are using Dell PowerEdge
servers.
For other types of servers, please refer to their manuals.

### \[Informational\] Update steps for Dell PowerEdge servers

1. Download the latest stable iDRAC and BIOS firmware from Dell's site.

2. Update iDRAC and BIOS via the iDRAC Web interface.

3. Execute "Full Power Cycle" via the iDRAC Web interface.


Collect Redfish Data
--------------------

You need to collect Redfish data from your servers.
Prepare a `collector` binary file and its [configuration files](config.md)
on a server, and execute:

```console
$ collector show > data.json
```

This saves collected Redfish data into a file in a [simple JSON format](collector.md#data-format).
The rest of works are performed on the retrieved JSON file, and can be run
on your local machine.

`collector` retrieves Redfish data by traversing from `/redfish/v1`
by default.
If your server's BMC provides Redfish API at other than `/redfish/v1`,
prepare a base rule file with `Traverse.Root` filled, and use it by
`--base-rule=<file>`.

```console
$ cat rule.yaml
Traverse:
  Root: /my/redfish/v1

$ collector show --base-rule=rule.yaml > data.json
```


Summarize Redfish Data
----------------------

Next you should find informative properties to be exported from the collected
Redfish data.
The collected data will be very large, however.
For example our Dell server returns 1.8 MB data.
Before looking closely at the data, you can condense the data.

The following command shows the collected data as is.

```console
$ collector show --input-file=data.json
```

Redfish API returns one page of some hardware information in a JSON format
per access, according to the accessed URL path.
For example, `GET /redfish/v1` returns a page of the server's overall
information, while `GET /redfish/v1/Systems/system1/Storage/Volumes/volume1`
returns a page of very specific volume's information.
`collector show` traverses such pages and displays them in the form of
a JSON key-value object.
The keys of the JSON object are paths of Redfish data, and the values are
pages returned by Redfish API in JSON.
See the ["Data Format" section](collector.md#data-format) to view an example.

### Exclude paths

If you find that Redfish pages at paths of `/redfish/v1/JSONSchemas/...`
or something do not provide useful data for metrics at all, you can exclude
those pages by specifying `Traverse.Excludes` in a base rule file.

```console
$ collector show --input-file=data.json --base-rule=rule.yaml
{
    "/redfish/v1/JSONSchemas": { ...(non-metric values)... },
    "/redfish/v1/Chassis/chassis1/Power": { ...(metric values)... }
}

$ vi rule.yaml
Traverse:
  Root: /redfish/v1
  Excludes:
    - /JSONSchemas

$ collector show --input-file=data.json --base-rule=rule.yaml
{
    "/redfish/v1/Chassis/chassis1/Power": { ...(metric values)... }
}
```

Each element of `Traverse.Exludes` are interpreted as a regular expression
and matched to Redfish paths.

The `--paths-only` option helps you to find unnecessary pages.

### Summarize pages with similar paths

Because Redfish data are well-structured, you will find that Redfish pages
with similar contents have similar paths like these:

```console
$ collector show --input-file=data.json --base-rule=rule.yaml
{
    "/redfish/v1/Chassis/chassis1/Power": { ...(metric values on chassis#1)... },
    "/redfish/v1/Chassis/chassis2/Power": { ...(metric values on chassis#2)... },
    "/redfish/v1/Chassis/chassis3/Power": { ...(metric values on chassis#3)... }
}
```

The first page would show the information about the power supply modules of
the chassis #1.
The second page would be about those of the chassis #2.
Such pages have the same structure in most cases.
In process of finding informative properties, only one of the similar pages
has importance.

`collector` can drop the second and later pages from the similar pages with
similar paths, but `collector` cannot detect "similar" paths by itself.
Instead, please specify `Metrics.Path` for "similar" paths in a base rule file.
Replace the varying path components in similar paths with a "pattern"
component like `{foo}`.
In the above example, you should specify `/redfish/v1/Chassis/{chassis}/Power`
in a base rule file.

```console
$ vi rule.yaml
Traverse:
  Root: /redfish/v1
Metrics:
  - Path: /redfish/v1/Chassis/{chassis}/Power

$ collector show --input-file=data.json --base-rule=rule.yaml
{
    "/redfish/v1/Chassis/chassis1/Power": { ...(metric values on chassis#1)... }
}
```

This will show at most one page per patterned path.
See the ["Patterned path" section](rule.md#patterned-path) for more detail on
`Metrics.Path`.

Note that specifying `Metrics.Path` here does not prohibit the data outside
the path to be shown.

Also note that specifying `Metrics.Path` here does not affect whether
properties inside the path will finally be collected and exported as metrics
or not.
It controls summarization in showing Redfish data.
It also controls the output format of `Metrics.Path` in the `generate-rule`
mode, but whether the path is included in the generated rule depends on
the user-specified options.

The `--paths-only` option helps you to find pages with similar paths.

### Summarize array elements

The JSON data in a Redfish page sometimes contains arrays to represent groups
of resources.
For example, `PowerSupplies` in `/redfish/v1/Chassis/chassis1/Power` is
an array of information on power supply units.

```console
$ collector show --input-file=data.json
{
    "/redfish/v1/Chassis/chassis1/Power": {
        "PowerSupplies": [
            { ...(metric values on PSU#1)... },
            { ...(metric values on PSU#2)... },
            { ...(metric values on PSU#3)... }
        ]
    }
}
```

In process of finding informative properties, only one of the array elements
has importance.
By specifying the `--truncate-arrrays` option, you can drop the second and
later array elements.

```console
$ collector show --input-file=data.json --truncate-arrays
{
    "/redfish/v1/Chassis/chassis1/Power": {
        "PowerSupplies": [
            { ...(metric values on PSU#1)... }
        ]
    }
}
```

This will show at most one element per array.

### Condense more

The `--omit-empty` option truncates empty arrays and empty objects in
Redfish pages.

The `--ignore-field` option hides those key-value pairs in Redfish pages
whose keys match the specified pattern.
For example it would be comfortable to hide Redfish meta data by specifying
`--ignore-field="@odata.*"`.


Select Key Properties
---------------------

Now you have much smaller data.
You will be able to find informative properties from the data.

```console
$ collector show --input-file=data.json ...
{
  ...
  "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1": {
    ...
    "Status": {
      "Health": "OK",
      "State": "Enabled"
    },
    ...
  },
  ...
}
```

You need to record the names and "types" of informative properties.
Types are the names of converting rules for Redfish data to be exported
as metrics numbers.
The ["Type of property" section](rule.md#type-of-property) describes
currently supported types.
If you cannot find an appropriate type, please implement a new one in
[converter.go](../redfish/converter.go).

In the example above, two properties seem informative; "Health" of
the type of `health` and "State" of the type of `state`.

Once you find an informative property, you can specify `--ignore-field`
with it to reduce data before finding other informative properties.

```console
$ collector show --input-file=data.json ... --ignore-field=Health --ignore-field=State
(...much smaller data...)
```

Generate a Rule File
--------------------

Now you have a set of name-type pairs of informative properties.
Let's generate a rule file by specifying them in the generate mode of
`collector`.

```console
$ collector generate-rule --base-rule=rule.yaml --key=Health:health --key=State:state data.json
```

Check the generated file carefully.
If you find unnecessary paths or redundant rules, go back to summarization
of the data.

You may need to fix auto-generated `Metrics.Properties.Pointer` and
`Metrics.Properties.Name`.
Especially pattern names in `Metrics.Properties.Pointer` are used as
label names in Prometheus metrics, so please take a good look.
See the ["Patterned pointer" section](rule.md#patterned-pointer) for more
detail on `Metrics.Properties.Pointer`.


Generate a Common Rule File for Multiple Types of Machines
----------------------------------------------------------

If you have serveral types of machines with a little variation, first perform
the collect, summarize, and select steps for each type.
You will be able to reuse summarizing options.
After you find informative properties, you can generate a common rule file
by passing all collected Redfish data to one command invocation.

```console
$ collector generate-rule --base-rule=rule.yaml --key=Health:health --key=State:state data1.json data2.json data3.json
```


Put a Rule File into Rules Directory
------------------------------------

To use the generated rule file, put it in [redfish/rules](../redfish/rules)
with an appropriate name.
See [monitor-hw's root.go](../pkg/monitor-hw/cmd/root.go) to learn about
loading of a rule file.
Update root.go's logic if necessary.

Currently `monitor-hw` works as follows:
1. Detects the hardware vendor by reading `/sys/devices/virtual/dmi/id/sys_vendor`.
2. Detects the Redfish version by retrieving `/redfish/v1` from the Redfish API
   and inspecting the `RedfishVersion` property.
3. Constructs a file name as `dell_redfish_{version}.yml` for Dell servers.
   Other vendors are not supported.


[Prometheus]: https://prometheus.io/
[Redfish]: https://www.dmtf.org/standards/redfish
