# Change Log
## 1.3.5 (2022-06-20)

### Enhancements
- Use tooltip UI for function control popup

### Bug fixes
- **Backend**: Fix "json: cannot unmarshal" error

## 1.3.4 (2022-04-13)

### Breaking changes
- Grafana v7.X.X or earlier is not supported from this version

### Enhancements
- Build system for documentaion uses myst-parser instead of commonmark

### Bug fixes
- **Frontend**: Fix initial value of function add button

## 1.3.3 (2021-10-15)

### Enhancements
- **Backend**: Add pvname label to timeseries dataframe for multi-dimensional alerting

### Bug fixes
- **Backend**: Fix dataframe name from response to its PV name
- **Backend**: Add display name to dataframe config to properly display legend
- **Frontend**: Fix function adding button

## 1.3.2 (2021-04-09)

### Enhancements
- **Backend / Frontend**: Support PV name with recursive alternation
- **Backend**: Support Extrapolation for visualization

### Bug fixes
- **Backend**: `Delta` function faulted on 1-length datasets

## 1.3.1 (2021-03-23)

### Features / Enhancements
- Reduce query size when using backend data source for visualization
- Support alternation pattern like `(pv:1|pv:2):other` with backend data source

### Bug fixes
- Fix stream parameters to be able to use variables
- Fix the problem that all plots were displayed with the name "values" when using the back end data source 

## 1.3.0 (2021-03-05)

### Features / Enhancements
- Add stream feature for live updating
- Add Go Backend and alerts feature
- Add "Use Backend" button for configuration
- Add help text for the query fields

### Bug fixes
- Fix text width of Function parmeter during focusing on it

## 1.2.0 (2020-10-30)

### Features / Enhancements
- Support array data 
- Add `Array to Scalar` functions
- Add `movingAverage` function
- Improve query process to retrieve data from Archiver Appliance efficiently

## 1.1.1 (2020-10-23)

### Features / Enhancements
- Add disableExtrapol function
- Fix testDatasource for datasource test
- Change retrieving data format from json to qw

## 1.1.0 (2020-10-19)

### Features
- Add extrapolation feature for raw operator
- Add disableAutoRaw function
- Add sort functions
- Add `last` operator

## 1.0.1 (2020-06-19)

### Bug Fixes
- Add name field to DataFrame to use PV name in Data link

## 1.0.0 (2020-06-05)

### Breaking changes
- Grafana v6.X.X or earlier is not supported from this version
- Installation and update process is changed since `dist` folder was removed from master branch

### Features / Enhancements
- Support Grafana v7.0.0 or later
- Development process is changed with [grafana-toolkit](https://github.com/grafana/grafana/tree/master/packages/grafana-toolkit)
- Source codes are migrated from JavaScript to TypeScript
- UI library is migrated from AngularJS to React
- Support variables for alias field and [Functions](functions) parameter
- Support limit parameter for variable query

## 0.1.2 (2020-04-14)

### Features
- Add binInterval option function to set a binning interval for processing of data
- Add exclude option to exclude matched PV data

## 0.1.1 (2020-04-07)

### Features
- Support variables for `Operator` field
- Add maxNumPVs option function to set a maximum number of PVs for a target

## 0.1.0 (2019-12-08)

### Features
- Select multiple PVs by using Regex (Only supports wildcard pattern like `PV.*` and alternation pattern like `PV(1|2)`)
- Legend alias with regular expression pattern
- Data retrieval with data processing (See [Archiver Appliance User Guide](https://slacmshankar.github.io/epicsarchiver_docs/userguide.htm) for processing of data)
- Using PV names for Grafana variables
- Transform your data with processing functions
