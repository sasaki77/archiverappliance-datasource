# Installation

This plugin is not registered in Grafana Plugin Repository.
Please put this plugin into your grafana plugins directory;
the default path is `/var/lib/grafana/plugins`.

## Install using releases

1. Download a pre-built release from the [releases page](https://github.com/sasaki77/archiverappliance-datasource/releases).
2. Unzip the contents in the Grafana plugins folder. The default path is `/var/lib/grafana/plugins`.
3. This plugin is unsigned. It must be specially listed by name in the Grafana `grafana.ini` file to allow Grafana to use it. Add `sasaki77-archiverappliance-datasource` to the `allow_loading_unsigned_plugins` parameter in the `[plugins]` section.

To update the plugin, follow the same step and remove the previous version.

## Install using git
We decided not to provide a git branch for a installation from version 1.3.0.

To install using git, please refer the development page.