# Tips
## Binning interval
Binning interval is used for processing of data except for raw operator.
Basically, the interval is determined by width of panel and time range.

For example if width of panel is 300 px and time range is 1 hour (= 3600 seconds) then interval is 12 seconds.

## Auto raw operation
When a binning interval is less than 1 second, raw data is retrieved even if other operator is set.
To disable auto raw operation, [disableAutoRaw](functions.html#disableAutoRaw) function is available.

## Extrapolation for raw operation
In raw operator mode, the retrieved data is extrapolated by last data value. 
This feature prevents a graph from cutting of or not being desplayed.
To disable the extrapolation, [disableExtrapol](functions.html#disableExtrapol) function is available.

## Array data support
The plugin supports array data from version 1.2.0.

Array data is managed like a table in the plugin. Following is a example of array data that has 3 elements:

```eval_rst
+---------------------+------------+------------+------------+  
| time                | PV:NAME[0] | PV:NAME[1] | PV:NAME[2] |
+=====================+============+============+============+  
| 2020-01-02 03:04:00 | 1          | 2          | 3          |
+---------------------+------------+------------+------------+  
| 2020-01-02 03:05:00 | 4          | 5          | 6          |
+---------------------+------------+------------+------------+  
| 2020-01-02 03:06:00 | 7          | 8          | 9          |
+---------------------+------------+------------+------------+  
```

The data has a column for `time` and columns for each array elements. Column name for a array element is set as PV name with suffix for an index.
This table format called [wide format data frame](https://grafana.com/docs/grafana/latest/developers/plugins/data-frames/#wide-format) is available from most of Grafana panel.

The plugin also supports conversion from array data to scalar data with the [Array to Scalar](functions.html#array-to-scalar-functions) functions.
Above array data becomes as following with [toScalarByMin](functions.html#toscalarbymin):

```eval_rst
+---------------------+---------------+
| time                | PV:NAME (min) |
+=====================+===============+
| 2020-01-02 03:04:00 | 1             |
+---------------------+---------------+
| 2020-01-02 03:05:00 | 4             |
+---------------------+---------------+
| 2020-01-02 03:06:00 | 7             |
+---------------------+---------------+
```

You can use as many `Array to Scalar` functions as you need for same array data.
The other functions except `Options` are not supported for array data.