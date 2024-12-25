## ybm cluster db-audit-logging describe

Describe Database Audit Logging configuration

### Synopsis

Describe Database Audit Logging configuration

```
ybm cluster db-audit-logging describe [flags]
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster db-audit-logging](ybm_cluster_db-audit-logging.md)	 - Configure Database Audit Logging for your Cluster.

