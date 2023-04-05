## ybm backup

Manage backup operations of a cluster

### Synopsis

Manage backup operations of a cluster

```
ybm backup [flags]
```

### Options

```
  -h, --help   help for backup
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!
* [ybm backup create](ybm_backup_create.md)	 - Create backup for a cluster in YugabyteDB Managed
* [ybm backup delete](ybm_backup_delete.md)	 - Delete backup for a cluster in YugabyteDB Managed
* [ybm backup get](ybm_backup_get.md)	 - Get list of existing backups available for a cluster in YugabyteDB Managed
* [ybm backup list](ybm_backup_list.md)	 - List existing backups available for a cluster in YugabyteDB Managed
* [ybm backup restore](ybm_backup_restore.md)	 - Restore backups into a cluster in YugabyteDB Managed

