# Development
## Development setup

This data source plugin consists of both frontend and backend components.

### Front-end Development

Node version v12.x is recommended. If you're new to the Node.js ecosystem, [Node Version Manager](https://github.com/nvm-sh/nvm) is a good place to start for managing different Node.js installations and environments.

[grafana-toolkit](https://grafana.com/developers/plugin-tools/) is used to develop the plugin. Please refer grafana-toolkit documentation for more information.

1. Begin by installing Yarn ([https://yarnpkg.com/](https://yarnpkg.com/))
    ```bash
    npm install -g yarn
    ```

2. Install dependencies
    ```bash
    yarn install
    ```

3. Build plugin in development mode or run in watch mode
    ```bash
    yarn dev
    ```
    or
    ```bash
    yarn watch
    ```

4. Build plugin in production mode
    ```bash
    yarn build
    ```

### Back-end Development

If the Mage build tool is not already installed, you may install it using the installation instructions on the [Mage homepage](https://magefile.org/)

1. Update [Grafana plugin SDK for Go](https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/grafana-plugin-sdk-for-go) dependency to the latest minor version:

    ```bash
    go get -u github.com/grafana/grafana-plugin-sdk-go
    ```

2. Build backend plugin binaries for Linux, Windows and Darwin:
    ```bash
    mage -v
    ```

3. List all available Mage targets for additional commands:
    ```bash
    mage -l
    ```

## Build documentation
This documentation is build by Sphinx. Follow below steps to build documentation.

```
python -m venv env
source env/bin/activate
pip install sphinx myst-parser sphinx_rtd_theme
make html
```
