## ybm backup list

List existing backups available for a cluster in YugabyteDB Aeon

### Synopsis

List existing backups available for a cluster in YugabyteDB Aeon

```
ybm backup list [flags]
```

### Options

```
      --cluster-name string   [OPTIONAL] Name of the cluster to fetch backups.
  -h, --help                  help for list
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm backup](ybm_backup.md)	 - Manage backup operations of a cluster

