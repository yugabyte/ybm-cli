## ybm cluster db-query-logging

Configure Database Query Logging for your Cluster.

### Synopsis

Configure Database Query Logging for your Cluster.

```
ybm cluster db-query-logging [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for db-query-logging
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
* [ybm cluster db-query-logging describe](ybm_cluster_db-query-logging_describe.md)	 - Describe Database Query Logging config
* [ybm cluster db-query-logging disable](ybm_cluster_db-query-logging_disable.md)	 - Disable Database Query Logging
* [ybm cluster db-query-logging enable](ybm_cluster_db-query-logging_enable.md)	 - Enable Database Query Logging
* [ybm cluster db-query-logging update](ybm_cluster_db-query-logging_update.md)	 - Update Database Query Logging config

