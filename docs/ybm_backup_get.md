## ybm backup get

Get list of existing backups available for a cluster in YugabyteDB Managed

### Synopsis

Get list of existing backups available for a cluster in YugabyteDB Managed

```
ybm backup get [flags]
```

### Options

```
      --cluster-name string   [OPTIONAL] Name of the cluster to fetch backups.
  -h, --help                  help for get
```

### Options inherited from parent commands

```
  -a, --apiKey string     YBM Api Key
      --config string     config file (default is $HOME/.ybm-cli.yaml)
      --debug             Use debug mode, same as --logLevel debug
  -l, --logLevel string   Select the desired log level format(info). Default to info
      --no-color          Disable colors in output , default to false
  -o, --output string     Select the desired output format (table, json, pretty). Default to table
      --wait              Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm backup](ybm_backup.md)	 - Manage backup operations of a cluster

