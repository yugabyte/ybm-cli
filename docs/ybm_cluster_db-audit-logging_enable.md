## ybm cluster db-audit-logging enable

Enable Database Audit Logging

### Synopsis

Enable Database Audit Logging

```
ybm cluster db-audit-logging enable [flags]
```

### Options

```
      --integration-name string      [REQUIRED] Name of the Integration
      --ysql-config stringToString   [REQUIRED] The ysql config to setup DB audit logging
                                     	Please provide key value pairs as follows:
                                     	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
                                     	log_relation=<boolean>,log_statement_once=<boolean> (default [])
      --statement_classes string     [REQUIRED] The ysql config statement classes
                                     	Please provide key value pairs as follows:
                                     	statement_classes=READ,WRITE,MISC
  -h, --help                         help for enable
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

