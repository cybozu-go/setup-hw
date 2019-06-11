Collect Redfish Data
====================

`collector` is a command-line tool that collects [Redfish][] data for creating/updating [collection rules](rule.md).

Synopsis
--------

```console
$ collector show [--input-file=<file>]
$ collector list-status [--input-file=<file>]
$ collector generate-rule [--input-file=<file>] [--key=<key>...]
```

Description
-----------

`collector` is a command-line tool that collects and processes [Redfish][] data.
It shows the whole data, lists paths which hold statuses, or generates a rule files.

In any mode, `collector` uses pre-collected Redfish data if `--input-file` is specified,
or collects Redfish data by accessing Redfish API if `--input-file` is not specified.

#### show whole data

`collector show` outputs Redfish data in the format described below.

#### list status paths

`collector list-status` outputs Redfish data locations where `Status` keys are returned.
The output is formatted as a JSON object whose keys are URL paths and whose values are lists of JSON Pointers each of which points `Status`.

```json
{
  "/redfish/v1/Chassis/Chassis1/Power": ["/Status", "/Voltages/0/Status", "/Voltages/1/Status"]
}
```

#### generate a rule file

`collector generate-rule` outputs a [collection rule](rule.md) to collect specified keys as metrics.

`collector generate-rule` cannot detect [patterns](rule.md#patterned-path) in paths, i.e. it cannot generate rules with page paths like `/redfish/v1/Chassis/{chassis}`.
it just lists all page paths.

`collector generate-rule` summarizes keys if they are in a list.
It generates `Pointer` of `/foo/{TBD}/Status` if it finds `Status` keys in the list at `/foo`.
This is based on [patterns](rule.md#patterned-pointer) in a collection rule.

Options
-------

`--input-file=<file>` specifies the path of the input data file.
`collector` accesses Redfish on the local machine without this option.
The format of the input data file is same as that of the output of `collector show`.

`--key=<key>` specifies the key to find in generating a rule file.
This can be specified in the `generate-rule` mode only.
This option can be specified multiple times.

Data Format
-----------

`collector show` outputs collected Redfish data in the following JSON format.

- key: path of Redfish data
- value: JSON data retrieved by accessing Redfish API

The output is like this:

```json
{
    "/redfish/v1/Chassis/": {
        "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
        "@odata.id": "/redfish/v1/Chassis/",
        "@odata.type": "#ChassisCollection.ChassisCollection",
        "Description": "Collection of Chassis",
        "Members": [
            {
                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
            }
        ],
        "Members@odata.count": 1,
        "Name": "Chassis Collection"
    },
    "/redfish/v1/Chassis/Chassis1": {
        ...
    },
    ...
}
```

Configuration files
-------------------

`collector` reads the host and the user information of the BMC from
the configuration files.
See the [description of configuration files](config.md) for details.
