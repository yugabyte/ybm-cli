## ybm backup policy

Manage backup policy of a cluster

### Synopsis

Manage backup policy of a cluster

```
ybm backup policy [flags]
```

### Options

```
  -h, --help   help for policy
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon Api Key
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
* [ybm backup policy disable](ybm_backup_policy_disable.md)	 - Disable backup policies
* [ybm backup policy enable](ybm_backup_policy_enable.md)	 - Enable backup policies
* [ybm backup policy list](ybm_backup_policy_list.md)	 - List backup policies
* [ybm backup policy update](ybm_backup_policy_update.md)	 - Update backup policies

