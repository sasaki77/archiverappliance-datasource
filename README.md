# EPICS Archiver Appliance Datasource

Visualize EPICS Archiver Appliance on Grafana.

![Screenshot](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-query-overview.png)


See [Archiver Appliance site](https://slacmshankar.github.io/epicsarchiver_docs/index.html) for more information about Archiver Appliance.

## Features
- Select multiple PVs by using Regex (Only supports wildcard pattern like `PV.*` and alternation pattern like `PV(1|2)`)
- Legend alias with regular expression pattern
- Data retrieval with data processing (See [Archiver Appliance User Guide](https://slacmshankar.github.io/epicsarchiver_docs/userguide.html) for processing of data)
- Using PV names for Grafana variables
- Transform your data with processing functions

## Installation

Clone this plugin into grafana plugins directory; the default is /var/lib/grafana/plugins.

## Documentaion
Documentation is available at [https://sasaki77.github.io/archiverappliance-datasource](https://sasaki77.github.io/archiverappliance-datasource).

## Configuration

This section lists the available configuration options for the Archiver Appliance data source.

### Query editor

| Configuration | Description |
|---------------|-------------|
| **PV** | Set PV name to be visualized. It is allowed to set multiple PVs by using Regular Expressoins alternation pattern (e.g. `(PV:1\|PV:2)`). |
| **Regex** | Enable/disable Regex mode. Refer [Select Multiple PVs by Regex](#select-multiple-pvs-by-regex). |
| **Operator** | Controls processing of data during data retrieval (Default: `mean`). Refer [Archiver Appliance User Guide](https://slacmshankar.github.io/epicsarchiver_docs/userguide.html) about processing of data. Special operator `raw` and `last` are also available. `raw` allows to retrieve the data without processing. `last` allows to retrieve the last data in the specified time range. |
| **Alias** | Set alias for legend. |
| **Alias pattern** | Set regular expressoin pattern to use PV name for legend alias. Refer [Legend Alias with Regex Pattern](#legend-alias-with-regex-pattern) |
| **Function** | Apply processing function for retrieved data. Refer [Apply Processing Functions](#apply-processing-functions) |

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

It is recommend to use node 12.x.

```
npm install -g yarn
yarn install
yarn dev
```

[grafana-toolkit](https://github.com/grafana/grafana/tree/master/packages/grafana-toolkit) is used to develop the plugin. Please refer grafana-toolkit documentation for more information.

### Test environment with Docker Compose
Thanks to [pklaus / docker-archiver-appliance](https://github.com/pklaus/docker-archiver-appliance) and [pklaus / archiver-appliance-with-example-ioc](https://github.com/pklaus/archiver-appliance-with-example-ioc), the test environment is available with Docker Compose.

```bash
docker-compose up
```

The following containers are runinng after `docker-compse up`.

| Name | Description |
|---------------|-------------|
| **grafana** | Runs a Grafana server. |
| **archappl** | Runs a EPICS Archiver Appliance. |
| **redis** | Runs a datastore for the persistance of the appliance configuration. |
| **example** | Runs a example EPICS IOC to be archived. |

To set up the Archiver Appliance, open http://localhost:17665/mgmt/ui/index.html. You can add the PVs served by the example IOC on this page.
Enter the following lines in the input field and then clicking the `Archive` button.
the Archive will be start a few minuites later.

```
root:circle:tick
root:circle:step
root:circle:period
root:line:b
root:aiExample
root:aiExample1
root:ai1
root:aiExample2
root:ai2
root:aiExample3
root:ai3
root:EXAMPLE:version
root:compressExample
root:circle:angle
root:line:a
root:circle:x
root:circle:y
root:calcExample
root:calcExample1
root:calc1
root:calcExample2
root:calc2
root:calcExample3
root:calc3
```

![Archiver-Appliance-mgmt-page](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-test-mgmt-page.png)

To add a data source, open Grafana (http://localhost:3000). On the data sources page, add a data source and set URL as http://archappl:17665/retrieval.

![test-configuration](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-test-configuration.png)

The following is a example query with this test environment.

![test-query](https://sasaki77.github.io/archiverappliance-datasource/_images/aa-test-query.png)

## Build documentation

```
python -m venv env
source env/bin/activate
pip install sphinx commonmark recommonmark sphinx_rtd_theme
make html
```
