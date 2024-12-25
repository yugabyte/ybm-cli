## ybm cluster db-audit-logging

Configure Database Audit Logging for your Cluster.

### Synopsis

Configure Database Audit Logging for your Cluster.

```
ybm cluster db-audit-logging [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for db-audit-logging
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm cluster db-audit-logging describe](ybm_cluster_db-audit-logging_describe.md)	 - Describe Database Audit Logging configuration
* [ybm cluster db-audit-logging disable](ybm_cluster_db-audit-logging_disable.md)	 - Disable Database Audit Logging
* [ybm cluster db-audit-logging enable](ybm_cluster_db-audit-logging_enable.md)	 - Enable Database Audit Logging
* [ybm cluster db-audit-logging update](ybm_cluster_db-audit-logging_update.md)	 - Update Database Audit Logging Configuration

