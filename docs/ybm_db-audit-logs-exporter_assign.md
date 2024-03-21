## ybm db-audit-logs-exporter assign

Assign DB Audit

### Synopsis

Assign DB Audit Logs to a Cluster

```
ybm db-audit-logs-exporter assign [flags]
```

### Options

```
      --integration-id string        [REQUIRED] The ID of the Integration
      --ysql-config stringToString   [REQUIRED] The ysql config to setup DB auditting
                                     	Please provide key value pairs as follows:
                                     	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
                                     	log_relation=<boolean>,log_statement_once=<boolean> (default [])
      --statement_classes string     [REQUIRED] The ysql config statement classes
                                     	Please provide key value pairs as follows:
                                     	statement_classes=READ,WRITE,MISC
      --cluster-id string            [REQUIRED] The cluster ID to assign DB auditting
  -h, --help                         help for assign
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

* [ybm db-audit-logs-exporter](ybm_db-audit-logs-exporter.md)	 - Manage DB Audit Logs

