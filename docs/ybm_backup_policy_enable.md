## ybm backup policy enable

Enable backup policies

### Synopsis

Enable backup policies for cluster in YugabyteDB Managed

```
ybm backup policy enable [flags]
```

### Options

```
      --cluster-name string   [REQUIRED] Name of the cluster to enable backup policies.
  -h, --help                  help for enable
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YBM Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm backup policy](ybm_backup_policy.md)	 - Manage backup policy of a cluster

