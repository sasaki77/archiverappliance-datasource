# EPICS Archiver Appliance Datasource

Visualize EPICS Archiver Appliance on Grafana. This plugin is based on [simple-json-datasource](https://github.com/grafana/simple-json-datasource)

See [Archiver Appliance site](https://slacmshankar.github.io/epicsarchiver_docs/index.html) for more information.

## Installation

Clone this plugin into grafana plugins directory; the default is /var/lib/grafana/plugins.

## Dev setup

This plugin requires node 6.10.0

```
npm install -g yarn
yarn install
npm run build
```
