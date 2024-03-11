## ybm db-audit-logs-exporter unassign

Unassign DB Audit Logs Export Config

### Synopsis

Unassign DB Audit Logs Export Config

```
ybm db-audit-logs-exporter unassign [flags]
```

### Options

```
      --export-config-id string   [REQUIRED] The ID of the DB audit export config
      --cluster-id string         [REQUIRED] The cluster ID to assign DB auditting
  -f, --force                     Bypass the prompt for non-interactive usage
  -h, --help                      help for unassign
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

