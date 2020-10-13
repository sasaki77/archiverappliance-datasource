# Installation

This plugin is not registered in Grafana Plugin Repository.
Please put this plugin into your grafana plugins directory;
the default path is `/var/lib/grafana/plugins`.

## Install with git

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

### Update
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

## Install with .tar.gz file or .zip file
Download the `.tar.gz` file or `.zip` file and extract it into your grafana plugins directory.

Download link is available on [GitHub release page](https://github.com/sasaki77/archiverappliance-datasource/releases).

For example, you can install v1.0.0 as following.

```
wget https://github.com/sasaki77/archiverappliance-datasource/archive/1.0.0.tar.gz
tar -zxvf 1.0.0.tar.gz
```

Restart Grafana.

```bash
systemctl restart grafana-server
```

To update the plugin, follow the same step and remove the previous version.