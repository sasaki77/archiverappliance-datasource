# Change Log

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
