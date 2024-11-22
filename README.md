# EPICS Archiver Appliance Datasource

Visualize EPICS Archiver Appliance on Grafana.

![Screenshot](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-query-overview.png)


See [Archiver Appliance site](https://epicsarchiver.readthedocs.io) for more information about Archiver Appliance.

## Features
- Select multiple PVs by using Regex (Only supports wildcard pattern like `PV.*` and alternation pattern like `PV(1|2)`)
- Legend alias with regular expression pattern
- Data retrieval with data processing (See [Archiver Appliance User Guide](https://epicsarchiver.readthedocs.io/en/latest/user/userguide.html#processing-of-data) for processing of data)
- Using PV names for Grafana variables
- Transform your data with processing functions
- Live update with stream feature
- Find and notify problems with alerting feature

## Installing the plugin

### Install with Grafana CLI
1. Install the plugin with Grafana CLI. Execute Grafana CLI as following:
```bash
# Install latest version. You can also use this command to update the plugin to the latest version.
grafana-cli --pluginUrl https://github.com/sasaki77/archiverappliance-datasource/releases/latest/download/archiverappliance-datasource.zip plugins install sasaki77-archiverappliance-datasource

# Install particular version. This example will install v1.4.2.
grafana-cli --pluginUrl https://github.com/sasaki77/archiverappliance-datasource/releases/download/1.4.2/archiverappliance-datasource.zip plugins install sasaki77-archiverappliance-datasource

# Install nightly builds.
grafana-cli --pluginUrl https://github.com/sasaki77/archiverappliance-datasource/releases/download/nightly/archiverappliance-datasource.zip plugins install sasaki77-archiverappliance-datasource
```
2. This plugin is unsigned. It must be specially listed by name in the Grafana `grafana.ini` file to allow Grafana to use it. Add `sasaki77-archiverappliance-datasource` to the `allow_loading_unsigned_plugins` parameter in the `[plugins]` section. See [Configure Grafana | Grafana documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/) for more detail on `grafana.ini`.

To update the plugin, execute Grafana CLI again.

### Install using Git

Clone this plugin into grafana plugins directory; the default is /var/lib/grafana/plugins.

### Configuring the plugin

Please consult [Configuration manual](https://sasaki77.github.io/archiverappliance-datasource/configuration.html) for instructions to configure the plugin following installation.

## Documentaion
Documentation is available at [https://sasaki77.github.io/archiverappliance-datasource](https://sasaki77.github.io/archiverappliance-datasource).

## Configuration

This section lists the available configuration options for the Archiver Appliance data source.

### Query editor

| Configuration     | Description                                                                                                                                                                                                                                                                                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **PV**            | Set PV name to be visualized. It is allowed to set multiple PVs by using Regular Expressoins alternation pattern (e.g. `(PV:1\|PV:2)`).                                                                                                                                                                                                                                                                      |
| **Regex**         | Enable/disable Regex mode. Refer [Select Multiple PVs by Regex](https://sasaki77.github.io/archiverappliance-datasource/query.html#select-multiple-pvs-by-regex).                                                                                                                                                                                                                                            |
| **Operator**      | Controls processing of data during data retrieval (Default: `mean`). Refer [Archiver Appliance User Guide](https://epicsarchiver.readthedocs.io/en/latest/user/userguide.html#processing-of-data) about processing of data. Special operator `raw` and `last` are also available. `raw` allows to retrieve the data without processing. `last` allows to retrieve the last data in the specified time range. |
| **Stream**        | Enable/Disable Stream mode. Stream allows to periodically update the data without refreshing the dashboard. The difference data from the last updated values is only retrieved.                                                                                                                                                                                                                              |
| **Interval**      | Streaming interval in milliseconds. You can also use a number with unit. e.g. `1s`, `1m`, `1h`. The default is determined by a width of panel and time range.                                                                                                                                                                                                                                                |
| **Capacity**      | The stream data is stored in a circular buffer. Capacity determines the buffer size. The default is detemined by a initial data size.                                                                                                                                                                                                                                                                        |
| **Alias**         | Set alias for legend.                                                                                                                                                                                                                                                                                                                                                                                        |
| **Alias pattern** | Set regular expressoin pattern to use PV name for legend alias. Refer [Legend Alias with Regex Pattern](https://sasaki77.github.io/archiverappliance-datasource/query.html#legend-alias-with-regex-pattern)                                                                                                                                                                                                  |
| **Function**      | Apply processing function for retrieved data. Refer [Apply Processing Functions](https://sasaki77.github.io/archiverappliance-datasource/query.html#apply-processing-functions)                                                                                                                                                                                                                              |

### Variables

[Variables](https://sasaki77.github.io/archiverappliance-datasource/variables.html) are supported for PV names registered in Archiver Appliance.

![Variables](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-variables-settings.png)

You can use regular expression pattern same as [Query Edit](query.html#select-multiple-pvs-by-regex).
Only wildcard pattern and alternation pattern are available.

You can also use another variable as a part of query.
For example, you have variable _group_, which is a list of PV name prefixes
according to PV naming rules on your institute.

Below is a query in this case.

```bash
${group}:.*
```

Default maximum number of PV names you can use in variables is 100. Limit parameter is available to change maxmum number. Parameters must be follow `?` character.

```bash
PV:NAME:.*?limit=1000
```

## Development setup

This data source plugin consists of both frontend and backend components.

### Front-end Development

Node version v12.x is recommended. If you're new to the Node.js ecosystem, [Node Version Manager](https://github.com/nvm-sh/nvm) is a good place to start for managing different Node.js installations and environments.

[grafana-toolkit](https://github.com/grafana/grafana/tree/master/packages/grafana-toolkit) is used to develop the plugin. Please refer grafana-toolkit documentation for more information.

1. Begin by installing Yarn (https://yarnpkg.com/)
```BASH
npm install -g yarn
```

2. Install dependencies
```BASH
yarn install
```

3. Build plugin in development mode or run in watch mode
```BASH
yarn dev
```
or
```BASH
yarn watch
```
4. Build plugin in production mode
```BASH
yarn build
```

### Back-end Development

If the Mage build tool is not already installed, you may install it using the installation instructions on the [Mage homepage](https://magefile.org/)

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

```bash
go get -u github.com/grafana/grafana-plugin-sdk-go
```

2. Build backend plugin binaries for Linux, Windows and Darwin:
```BASH
mage -v
```

3. List all available Mage targets for additional commands:
```BASH
mage -l
```

## Build documentation

```
python -m venv env
source env/bin/activate
pip install sphinx myst-parser sphinx_rtd_theme
make html
```

## Build Status

|       Build:        |                                                             Status:                                                              |
| :-----------------: | :------------------------------------------------------------------------------------------------------------------------------: |
|   Latest release    |       ![awaiting build](https://github.com/sasaki77/archiverappliance-datasource/workflows/publish-release.yml/badge.svg)        |
|    Master Branch    | ![awaiting build](https://github.com/sasaki77/archiverappliance-datasource/workflows/build-and-test.yml/badge.svg?branch=master) |
| Local Master Branch |                          ![awaiting build](../../workflows/build-and-test.yml/badge.svg?branch=master)                           |
| Local Latest Branch |                                 ![awaiting build](../../workflows/build-and-test.yml/badge.svg?)                                 |
