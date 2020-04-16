# Release Notes
## 0.1.3

**2020-XX-XX**

### Features
- Support variables for [Functions](functions) parameter.
- Support limit parameter for variable query

## 0.1.2

**2020-04-14**

### Features
- Add binInterval option function to set a binning interval for processing of data
- Add exclude option to exclude matched PV data

## 0.1.1

**2020-04-07**

### Features
- Support variables for `Operator` field
- Add maxNumPVs option function to set a maximum number of PVs for a target

## 0.1.0

**2019-12-08**

### Features
- Select multiple PVs by using Regex (Only supports wildcard pattern like `PV.*` and alternation pattern like `PV(1|2)`)
- Legend alias with regular expression pattern
- Data retrieval with data processing (See [Archiver Appliance User Guide](https://slacmshankar.github.io/epicsarchiver_docs/userguide.htm) for processing of data)
- Using PV names for Grafana variables
- Transform your data with processing functions
