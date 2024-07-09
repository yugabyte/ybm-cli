## ybm backup restore

Restore backups into a cluster in YugabyteDB Managed

### Synopsis

Restore backups into a cluster in  YugabyteDB Managed

```
ybm backup restore [flags]
```

### Options

```
      --backup-id string      [REQUIRED] ID of the backup to be restored.
      --cluster-name string   [REQUIRED] Name of the cluster to restore backups.
  -h, --help                  help for restore
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

