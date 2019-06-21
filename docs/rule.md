Redfish Data Collection Rule
============================

YAML files under [redfish/rules](../redfish/rules) are used by `monitor-hw`
as Redfish collection rules.
They instruct how to traverse [Redfish][] pages, how to interpret data,
and how to export them to [Prometheus][], one file for each hardware type.

They are not configuration files.
They are compiled into the `monitor-hw` binary.

Each file has the following top-level fields:

Name     | Required  | Type                  | Description
-------- | --------- | --------------------- | -----------
Traverse | true      | Traverse Rule         | See [Traverse Rule](#traverse-rule).
Metrics  | false (*) | array of Metric Rules | See [Metric Rule](#metric-rule).

* Though `Metrics` is marked as non-required, a rule with empty `Metrics`
produces no metrics.


Traverse Rule
-------------

A traverse rule specifies how to traverse Redfish pages.

Name     | Required | Type             | Description
-------- | -------- | ---------------- | -----------
Root     | true     | string           | Root path of Redfish, e.g. `/redfish/v1`.
Excludes | false    | array of strings | Path patterns in [regexp][] format which should not be traversed.


Metric Rule
-----------

Each metric rule corresponds to a page of Redfish data, or a set of
Redfish pages in the case of the patterned path, and specifies how to
find meaningful pages from traversed Redfish pages.

Name       | Required  | Type                   | Description
---------- | --------- | ---------------------- | -----------
Path       | true      | string                 | Path of a Redfish page.  [Patterned path](#patterned-path) can be used.
Properties | false (*) | array of Property Rule | See [Property Rule](#property-rule).

* Though `Properties` is marked as non-required, a rule with empty `Properties`
produces no metrics.

### Patterned path

A patterned path contains one or more special path elements.
Each special path element is surrounded by `{` and `}`, and matches
any path element of that place.
The matching adds labels to the Prometheus metrics in the matched pages.

For example, a patterned path `/redfish/v1/Chassis/{chassis}` matches
pages of `/redfish/v1/Chassis/Chassis.1` and `/redfish/v1/Chassis/Chassis.2`.
Each metric in `/redfish/v1/Chassis/Chassis.1` will have a label of
`chassis="Chassis.1"`.


Property Rule
-------------

Each property rule corresponds to a property in a Redfish page, or a set of
properties in the case of the patterned pointer, and specifies how to
interpret and export the properties as Prometheus metrics.

Name    | Required | Type   | Description
------- | -------- | ------ | -----------
Pointer | true     | string | Pointer to a property in the page.
Name    | true     | string | Base name of a metric converted from the property.  Used with the prefix of `hw_`.
Help    | false    | string | Help text.
Type    | true     | string | Type of a property.  This controls conversion from Redfish string to metric float.

`Pointer` can be given in the form of [JSON Pointer][RFC6901] with extension
of [patterns](#patterned-pointer).
The following are not supported:
  * reference to an array element, e.g. `/2` to point the third element in the array
  * escapes, e.g. `/foo~1bar` to point the object value with the key of `foo/bar`

### Patterned pointer

A patterned pointer contains one or more special pointer elements.
Each special pointer element is surrounded by `{` and `}`, and matches
any array element of that place.
The matching adds labels to the Prometheus metric of the matched property.

For example, a patterned pointer `/Temperatures/{sensor}/ReadingCelsius`
matches `24` and `42` in the following JSON data.

```
{
  "Temperatures": [
    {
      "ReadingCelsius": 24
    },
    {
      "ReadingCelsius": 42
    }
  ]
}
```

Two metrics are produced, one has a value of 24 and a label of `sensor="0"`,
and another has 42 and `sensor="1"`.
This is easy to understand by comparing the patterned pointer with a standard
JSON pointer `/Temperatures/0/ReadingCelsius`, which points to `24`.

Note that a pattern in a pointer does not match an object key.
For example, a patterned pointer `/Temperatures/{sensor}/ReadingCelsius`
does not produce any metrics from the following JSON data.

```
{
  "Temperatures": {
    "Sensor0": {
      "ReadingCelsius": 24
    }
  }
}
```

### Type of property

The following types are supported.
See [Redfish Resource and Schema Guide](https://www.dmtf.org/dsp/DSP2046)
for more detail.

* `number`: for generic numerical properties

* `health`: for "Health" property in "Status" type
  * `"OK"` => 0
  * `"Warning"` => 1
  * `"Critical"` => 2
  * `null` => -1

* `state`: for "State" property in "Status" type
  * `"Enabled"` => 0
  * `"Disabled"` => 1
  * `"Absent"` => 2
  * `"Deferring"` => 3
  * `"InTest"` => 4
  * `"Quiesced"` => 5
  * `"StandbyOffline"` => 6
  * `"StandbySpare"` => 7
  * `"Starting"` => 8
  * `"UnavailableOffline"` => 9
  * `"Updating"` => 10

* `bool`: for generic boolean properties
  * `false` => 0
  * `true` => 1
  * `null` => -1


[Redfish]: https://www.dmtf.org/standards/redfish
[Prometheus]: https://prometheus.io/
[regexp]: https://golang.org/pkg/regexp/
[RFC6901]: https://tools.ietf.org/html/rfc6901
