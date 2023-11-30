## ybm backup policy update

Update backup policies

### Synopsis

Update backup policies for cluster in YugabyteDB Managed

```
ybm backup policy update [flags]
```

### Options

```
      --cluster-name string                       [REQUIRED] Name of the cluster to update backup policies.
      --full-backup-frequency-in-days int32       [OPTIONAL] Frequency of full backup in days. (default 1)
      --full-backup-schedule-days-of-week int32   [OPTIONAL] Days of the week when the backup has to run. A comma separated list of the first two letters of the days of the week. Eg: 'Mo,Tu,Sa' (default 1)
      --full-backup-schedule-time string          [OPTIONAL] Time of the day at which the backup has to run. Please specify local time in 24 hr HH:MM format. Eg: 15:04
  -h, --help                                      help for update
      --retention-period-in-days int32            [REQUIRED] Retention period of the backup in days. (default 1)
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

* [ybm backup policy](ybm_backup_policy.md)	 - Manage backup policy of a cluster

