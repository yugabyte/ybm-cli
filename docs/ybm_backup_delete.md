## ybm backup delete

Delete backup for a cluster in YugabyteDB Managed

### Synopsis

Delete backup for a cluster in YugabyteDB Managed

```
ybm backup delete [flags]
```

### Options

```
      --backup-id string   [REQUIRED] The backup ID.
  -h, --help               help for delete
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

