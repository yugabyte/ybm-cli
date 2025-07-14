## ybm cluster pitr-config create

Create PITR Config for a cluster

### Synopsis

Create PITR Config for a cluster in YugabyteDB Aeon

```
ybm cluster pitr-config create [flags]
```

### Options

```
  -h, --help                      help for create
  -p, --pitr-config stringArray   [REQUIRED] Information for the PITR Configs to be created. All values are mandatory. Available options for namespace type are YCQL and YSQL. Retention period must be between 2 and 14 days. Please provide key value pairs namespace-name=<namespace-name>,namespace-type=<namespace-type>,retention-period-in-days=<retention-period-in-days>.
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

