# Variables
Variable is Grafana feature to create interactive and dynamic dashboards also known as Templating.
Refer [Grafana Documentation](https://grafana.com/docs/grafana/latest/variables/templates-and-variables/) for more information.

This plugin supports to use PV names registered in Archiver Appliance for variables.

## Adding a Variable
You can add a variable by following steps.

1. Open `Dashboard settings` from top header.
2. Select `Variables` from side menu.
3. Click `Add variable`.
4. Configure variable settings. Refer [Variable Settings](#variable-settings)
5. Click `Add`.

## Variable Settings
Query format is explained in this documentation.
See [Grafana Documentation](https://grafana.com/docs/grafana/latest/variables/templates-and-variables/#adding-a-variable)
for other settings.

![Variable Settings](./img/aa-variables-settings.png)

You can use regular expression pattern same as [Query Edit](query.html#select-multiple-pvs-by-regex).
Only wildcard pattern and alternation pattern are available.

You can also use another variable as a part of query.
For example, you have variable _group_, which is a list of PV name prefixes
according to PV naming rules on your institute.

Below is a query in this case.

```bash
${group}:.*
```

``` Note:: Default maximum number of PV names you can use in variables is **100**.
```

Limit parameter is available to change maxmum number. Parameters must be follow `?` character.

```bash
PV:NAME:.*?limit=1000
```

## Variables Usage
Variables is allowed to use in each field and [Functions](functions) parameter except for `alias pattern` field.

You can select multiple PVs from drop down menu with variables.

![Variables Usage](./img/aa-variables-usage.png)
