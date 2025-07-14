## ybm cluster pitr-config describe

Describe PITR Configs of a namespace in a cluster

### Synopsis

Describe PITR Configs of a namespace in a cluster in YugabyteDB Aeon

```
ybm cluster pitr-config describe [flags]
```

### Options

```
      --namespace-name string   [REQUIRED] Namespace to be restored via PITR Config.
      --namespace-type string   [REQUIRED] The type of the namespace. Available options are YCQL and YSQL
  -h, --help                    help for describe
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster pitr-config](ybm_cluster_pitr-config.md)	 - Manage Cluster PITR Configs

