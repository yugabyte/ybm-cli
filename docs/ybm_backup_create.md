## ybm backup create

Create backup for a cluster in YugabyteDB Aeon

### Synopsis

Create backup for a cluster in YugabyteDB Aeon

```
ybm backup create [flags]
```

### Options

```
      --cluster-name string      [REQUIRED] Name for the cluster.
      --description string       [OPTIONAL] Description of the backup.
  -h, --help                     help for create
      --retention-period int32   [OPTIONAL] Retention period of the backup in days. (Default: 1)
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

* [ybm backup](ybm_backup.md)	 - Manage backup operations of a cluster

