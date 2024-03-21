## ybm db-audit-logs-exporter update

Update DB Audit

### Synopsis

Update DB Audit Log Configuration for a Cluster

```
ybm db-audit-logs-exporter update [flags]
```

### Options

```
      --export-config-id string      [REQUIRED] The ID of the DB audit export config
      --integration-id string        [REQUIRED] The ID of the Integration
      --ysql-config stringToString   The ysql config to setup DB auditting
                                     	Please provide key value pairs as follows:
                                     	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
                                     	log_relation=<boolean>,log_statement_once=<boolean> (default [])
      --statement_classes string     The ysql config statement classes
                                     	Please provide key value pairs as follows:
                                     	statement_classes=READ,WRITE,MISC
      --cluster-id string            [REQUIRED] The cluster ID to assign DB auditting
  -h, --help                         help for update
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

