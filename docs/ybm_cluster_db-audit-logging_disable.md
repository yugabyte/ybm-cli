## ybm cluster db-audit-logging disable

Disable Database Audit Logging

### Synopsis

Disable Database Audit Logging, if enabled

```
ybm cluster db-audit-logging disable [flags]
```

### Options

```
  -f, --force   Bypass the prompt for non-interactive usage
  -h, --help    help for disable
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

