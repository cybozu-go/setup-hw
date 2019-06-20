Collect Redfish Data
====================

`collector` is a command-line tool that collects [Redfish][] data for creating/updating [collection rules](rule.md).

Synopsis
--------

```console
$ collector show [--input-file=<file>] [--base-rule=<file>] [--keys-only] [--omitempty] [--no-dup] [--ignore-field=<field>...] [--required-field=<field>...]
$ collector generate-rule [--base-rule=<file>] [--key=<key>:<type>...] INPUT_FILE...
```

Description
-----------

`collector` is a command-line tool that collects and processes [Redfish][] data.
It shows the whole data, or generates a rule file.

If `--base-rule` is specified, `collector` reads the base rule file and uses the traversal rules.
It starts traversal from `Traverse.Root`.
It excludes data whose paths are listed in `Traverse.Excludes`.
`collector` also uses `Metrics.Path` in the base rule file when in the generate mode.
See the "Generate mode" section below.

Show mode
---------

`collector show` outputs Redfish data in the format described in the "Data Format" section.

### Options

If `--input-file` is specified, it loads Redfish API responses from the file.

If `--keys-only` is specified, `collector show` shows only path.

If `--omitempty` is specified, `collector show` will not show an empty array or an empty map.

If `--no-dup` is specified, `collector show` will truncate the second and later elements in a list.
It will also truncate the second and later pages from those pages whose paths are matched to a certain `Metrics.Path` pattern in a base rule file.

If `--ignore-field` is specified, `collector show` will not show the field that matches the specified name.
This option can be specified for multiple times.

If `--required-field` is specified, `collector show` will show the JSON object that has the specified field.
This option can be specified for multiple times.

Generate mode
-------------

`collector generate-rule` outputs a [collection rule](rule.md) to collect and export specified keys as metrics.

`collector generate-rule` does not access Redfish API directly due to its slowness.
Instead `collector generate-rule` uses results of `collector show` as inputs.
It traverses Redfish data from the input file, picks up specified keys, and generates a collection rule as its output.

`collector generate-rule` can take multiple input files.
It generates multiple collection rules internally, and merges rules into one.
By accepting multiple input files, it can produce a unified collection rule for slightly varying types of machines.

`collector generate-rule` summarizes keys if they are in a list.
It generates `Pointer` of `/foos/{foo}/Status` if it finds a `Status` key in the first item of the list at `/foos`.
This is based on [patterns](rule.md#patterned-pointer) in a collection rule.

In contrast, `collector generate-rule` cannot detect [patterns](rule.md#patterned-path) in paths, i.e. it cannot generate rules with page paths like `/redfish/v1/Chassis/{chassis}` by itself.
It just lists all page paths in default.
If patterned `Metrics.Path`s are given through the `--base-rule` option, however, it can use the patterned paths to unify matched pages when generating rules.

### Options

`--key=<key>:<type>` specifies the property key to be searched in generating a rule file, followed by the type of the property.
This can be specified in the `generate-rule` mode only.
This option can be specified for multiple times.

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
