Collect Redfish Data
====================

`collector` is a command-line tool that collects [Redfish][] data for creating/updating [collection rules](rule.md).

Synopsis
--------

```console
$ collector show [--input-file=<file>] [--base-rule=<file>] [--keys-only] [--omitempty] [--no-dup] [--ignore-field=<field>...] [--required-field=<field>...]
$ collector generate-rule [--base-rule=<file>] [--key=<key>:<type>...] INPUT_FILE
```

Description
-----------

`collector` is a command-line tool that collects and processes [Redfish][] data.
It shows the whole data, or generates a rule files.

If `--base-rule` is specified, `collector` will exclude data included in `Traverse.Excludes`
and paths are unified according to patterns specified in `Metrics.Path`.

Show mode
---------

`collector show` outputs Redfish data in the format described below.

### Options

If `--input-file` is specified, it loads Redfish API responses from the file.

If `--keys-only` is specified, `collector show` shows only path.

If `--omitempty` is specified, `collector show` will not show an empty array or an empty map.

If `--ignore-field` is specified, `collector show` will not show the field that matches the specified name.

If `--required-field` is specified, `collector show` will show the JSON object that has the specified field.

Generate mode
-------------

`collector generate-rule` outputs a [collection rule](rule.md) to collect specified keys as metrics.

`collector generate-rule` cannot detect [patterns](rule.md#patterned-path) in paths, i.e. it cannot generate rules with page paths like `/redfish/v1/Chassis/{chassis}`.
it just lists all page paths.

`collector generate-rule` summarizes keys if they are in a list.
It generates `Pointer` of `/foo/{TBD}/Status` if it finds a `Status` key in the first item of the list at `/foo`.
This is based on [patterns](rule.md#patterned-pointer) in a collection rule.

### Options

`--key=<key>:<type>` specifies the property key to be searched in generating a rule file, followed by the type of the property.
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


[Redfish]: https://www.dmtf.org/standards/redfish
