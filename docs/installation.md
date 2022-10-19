# Installation

This plugin is not registered in Grafana Plugin Repository.
Please put this plugin into your grafana plugins directory;
the default path is `/var/lib/grafana/plugins`.

## Install with Grafana CLI
1. Install the plugin with Grafana CLI. Execute Grafana CLI as following:
```bash
# Install latest version. You can also use this command to update the plugin to the latest version.
grafana-cli --pluginUrl https://github.com/sasaki77/archiverappliance-datasource/releases/latest/download/archiverappliance-datasource.zip plugins install sasaki77-archiverappliance-datasource

# Install particular version. This example will install v1.4.2.
grafana-cli --pluginUrl https://github.com/sasaki77/archiverappliance-datasource/releases/download/1.4.2/archiverappliance-datasource.zip plugins install sasaki77-archiverappliance-datasource
```
2. This plugin is unsigned. It must be specially listed by name in the Grafana `grafana.ini` file to allow Grafana to use it. Add `sasaki77-archiverappliance-datasource` to the `allow_loading_unsigned_plugins` parameter in the `[plugins]` section. See [Configure Grafana | Grafana documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/) for more detail on `grafana.ini`.

To update the plugin, execute Grafana CLI again.

## Install using git
We decided not to provide a git branch for a installation from version 1.3.0.

To install using git, please refer the [development page](development).