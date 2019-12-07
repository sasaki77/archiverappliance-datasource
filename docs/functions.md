# Functions
## Categories
Functions are categorized into below groups.

- **Transform:** converts the timeseries datapoints from its datapoint values.
- **Filter Series:** picks up some series that meet certain condition.

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

```js
scale(50)
scale(-0.5)
```

Adds _delta_ to each datapoint.

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

```js
top(5, max)
top(10, avg)
```

### _bottom_
``` function:: bottom(number, value)
```

Returns bottom N series sorted by _value_.
Available _value_ is as following: _avg_, _min_, _max_, _absoluteMin_, _absoluteMax_, and _sum_.

```js
bottom(5, avg)
bottom(10, min)
```
