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