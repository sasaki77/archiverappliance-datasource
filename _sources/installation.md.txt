# Installation
This plugin is not registered in Grafana Plugin Repository.
Please clone this plugin into your grafana plugins directory;
the default path is `/var/lib/grafana/plugins`.

Clone the codes with tag specication from GitHub.

```bash
git clone -b <verion> https://github.com/sasaki77/archiverappliance-datasource
```

For example, you can install v1.0.0 as following.

```bash
git clone -b 1.0.0 https://github.com/sasaki77/archiverappliance-datasource
```

Restart Grafana after cloning.

```bash
systemctl restart grafana-server
```

``` Warning:: Master branch has no longer dist folder from v1.0.0. Therefore, you can't install it with master branch without build.
```

## Update
To update the plugin, fetch repository from GitHub and checkout with version tag.

```bash
git fetch
git checkout -b <version> <version>
```

Verion list is available with `git tag -l`.

```bash
$ git tag -l
0.1.0
0.1.1
0.1.2
1.0.0
```