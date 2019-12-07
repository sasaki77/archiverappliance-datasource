# EPICS Archiver Appliance Datasource

Visualize EPICS Archiver Appliance on Grafana.

See [Archiver Appliance site](https://slacmshankar.github.io/epicsarchiver_docs/index.html) for more information about Archiver Appliance.

## Features
- Select multiple PVs by using Regex (Only supports wildcard pattern like `PV.*` and alternation pattern like `PV(1|2)`)
- Legend alias with regex pattern
- Data retrieval with data processing (See [Archiver Appliance User Guide](https://slacmshankar.github.io/epicsarchiver_docs/userguide.htm) for processing of data)
- Using PV names for Grafana variables
- Transform your data with processing functions

## Installation

Clone this plugin into grafana plugins directory; the default is /var/lib/grafana/plugins.

## Documentaion
Documentation is available at [https://sasaki77.github.io/archiverappliance-datasource](https://sasaki77.github.io/archiverappliance-datasource)

## Dev setup

It is recommend to use node 12.13.0

```
npm install -g yarn
yarn install
npm run build
```
