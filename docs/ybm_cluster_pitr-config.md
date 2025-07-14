## ybm cluster pitr-config

Manage Cluster PITR Configs

### Synopsis

Manage Cluster PITR Configs

```
ybm cluster pitr-config [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for pitr-config
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm cluster pitr-config clone](ybm_cluster_pitr-config_clone.md)	 - Clone namespace via PITR Config for a cluster
* [ybm cluster pitr-config create](ybm_cluster_pitr-config_create.md)	 - Create PITR Config for a cluster
* [ybm cluster pitr-config delete](ybm_cluster_pitr-config_delete.md)	 - Delete PITR Config for a cluster
* [ybm cluster pitr-config describe](ybm_cluster_pitr-config_describe.md)	 - Describe PITR Configs of a namespace in a cluster
* [ybm cluster pitr-config list](ybm_cluster_pitr-config_list.md)	 - List PITR Configs for a cluster
* [ybm cluster pitr-config update](ybm_cluster_pitr-config_update.md)	 - Update PITR Config for a cluster

