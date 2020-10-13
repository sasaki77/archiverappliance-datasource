# Functions
## Categories
Functions are categorized into below groups.

- **Transform:** converts the timeseries datapoints from its datapoint values.
- **Filter Series:** picks up some series that meet certain condition.
- **Options:** adds option parameters.

## Transform Functions

### _scale_
``` function:: scale(factor)
```

Multiplies datapoint by _factor_ at each point.

Examples:

```js
scale(10)
scale(-0.5)
```

### _offset_
``` function:: offset(delta)
```

Adds _delta_ to each datapoint.

Examples:

```js
offset(50)
offset(-0.5)
```

### _delta_
``` function:: delta()
```

Calculates difference between datapoint and its before point at each point.

### _fluctuation_
``` function:: fluctuation()
```

Subtracts first datapoint value from each datapoint.

## Filter Series Functions

### _top_
``` function:: top(number, value)
```

Returns top N series sorted by _value_.
Available _value_ is as following: _avg_, _min_, _max_, _absoluteMin_, _absoluteMax_, and _sum_.

Examples:

```js
top(5, max)
top(10, avg)
```

### _bottom_
``` function:: bottom(number, value)
```

Returns bottom N series sorted by _value_.
Available _value_ is as following: _avg_, _min_, _max_, _absoluteMin_, _absoluteMax_, and _sum_.

Examples:

```js
bottom(5, avg)
bottom(10, min)
```

### _exclude_
``` function:: exclude(pattern)
```

Exclude PV data that match the regular expression.

Examples:

```js
exclude(PV[0-9])
```

## Options Functions
### _maxNumPVs_
``` function:: maxNumPVs(number)
```

Set maximum number of PVs you can select for a target.

Examples:

```js
maxNumPVs(500)
```

### _binInterval_
``` function:: binInterval(interval)
```

Set binning interval for processing of data.
For example, if binInterval is set 100 and operator in query is set mean then returned values are average values in 100 seconds wide bins.

Examples:

```js
binInterval(100)
```
