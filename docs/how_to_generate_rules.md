How to Generate Data Collection Rules
=====================================

[monitor-hw](monitor-hw.md) exports metrics for [Prometheus][] by accessing
[Redfish][] API and converting retrieved data according to [data collection
rules](rule.md).
It uses YAML files under [redfish/rules](../redfish/rules) as sources of rules.

You can write a new rule file for a new type of servers by yourself, or you
can use [collector](collector.md) command-line tool.
This document describes how to use `collector` to generate a rule file.


Collect Redfish Data
--------------------

First of all, you need to collect Redfish data from your servers.
Prepare a `collector` binary file and its [configuration files] on a server,
and execute:

```console
$ collector show > data.json
```

This saves collected Redfish data in a [simple JSON format](collector.md#data-format)
into a file.

`collector` retrieves Redfish data by traversing from `/redfish/v1`.
If your server's BMC provides Redfish API at other than `/redfish/v1`,
prepare a base rule file with `Traverse.Root` filled, and use it by
`--base-rule=<file>`.

```console
$ cat rule.yaml
Traverse:
  Root: /my/redfish/v1

$ collector show --base-rule=rule.yaml > data.json
```

The rest of works are performed on the retrieved JSON file, and can be run
on your local machine.


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

### Exclude paths

If you find that Redfish pages at paths of `/redfish/v1/JSONSchemas/...`
or something do not provide useful data for metrics at all, you can exclude
those pages by specifying `Traverse.Excludes` in a base rule file.

```console
$ cat rule.yaml
Traverse:
  Root: /redfish/v1
  Excludes:
    - /JSONSchemas

$ collector show --input-file=data.json --base-rule=rule.yaml
```

Each element of `Traverse.Exludes` are interpreted as a regular expression
and matched to Redfish paths.

The `--keys-only` option helps you to find unnecessary pages.

### Summarize repeated pages

If you find that a group of Redfish pages has a patterned path like
`/redfish/v1/Chassis/.../Power`, you can summarize such pages by specifying
`Metrics.Path` in a base rule file, and by using the `--no-dup` option.

```console
$ cat rule.yaml
Traverse:
  Root: /redfish/v1
Metrics:
  - Path: /redfish/v1/Chassis/{chassis}/Power

$ collector show --input-file=data.json --base-rule=rule.yaml --no-dup
```

This will show at least one page per patterned path.
Note that specifying `Metrics.Path` here does not prohibit the data outside
the path to be shown.

See the ["Patterned path" section](rule.md#patterned-path) for more detail on
`Metrics.Path`.

Also note that specifying `Metrics.Path` here does not affect whether
properties inside the path will finally be collected and exported as metrics
or not.
It just controls summarization in deciding target properties.

The `--keys-only` option helps you to find repeated pages.

### Summarize repeated list elements

By using the `--no-dup` option, you can also summarize list elements.

```console
$ collector show --input-file=data.json --no-dup
```

This will show at least one element per list.

### Condense more

The `--omitempty` option truncates empty arrays and empty maps.

The `--ignore-field` option hides those fields whose names match the specified pattern.


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
In the above example, two properties are extracted; "Health" of the type of
`health` and "State" of the type of `state`.
Types are the names of converting rules for Redfish data to be exported
as metrics numbers.
They are defined in [converter.go](../redfish/converter.go).
If you cannot find an appropriate converter, add a new one.

You can use the `--required-field` option to see what values appear
in the specified kind of field throughout the data.
In this case, you should not use other summarizing options.

When you find an informative property, you can specify `--ignore-field`
with it to reduce data before finding other informative properties.


Generate a Rule File
--------------------

Now you have a set of name-type pairs of informative properties.
Let's generate a rule file by specifying them in the generate mode of
`collector`.

```console
$ collector generate-rule --base-rule=rule.yaml --key=name1:type1 --key=name2:type2 data.json
```

Check the generated file carefully.
If you find unnecessary paths or redundant rules, go back to summarization
of the data.

To use the generated rule file, put it in [redfish/rules](../redfish/rules)
with an appropriate name.
See [monitor-hw's root.go](../pkg/monitor-hw/cmd/root.go) to know loading
of a rule file.
Update root.go's logic if necessary.


Generate a Common Rule File for Multiple Types of Machines
----------------------------------------------------------

If you have serveral types of machines with a little variation, first perform
the collect, summarize, and select steps for each type.
You will be able to reuse summarizing options.
After you find informative properties, you can generate a common rule file
by passing all collected Redfish data to one command invocation.

```console
$ collector generate-rule --base-rule=rule.yaml --key=name1:type1 --key=name2:type2 data1.json data2.json data3.json
```


[Prometheus]: https://prometheus.io/
[Redfish]: https://www.dmtf.org/standards/redfish

