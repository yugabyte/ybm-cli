## ybm cluster db-query-logging enable

Enable Database Query Logging

### Synopsis

Enable Database Query Logging

```
ybm cluster db-query-logging enable [flags]
```

### Options

```
      --integration-name string            [REQUIRED] Name of the Integration
      --debug-print-plan string            [OPTIONAL] Enables various debugging output to be emitted. (default "false")
      --log-min-duration-statement int32   [OPTIONAL] Duration(in ms) of each completed statement to be logged if the statement ran for at least the specified amount of time. Default -1 (log all statements). (default -1)
      --log-connections string             [OPTIONAL] Log connection attempts. (default "false")
      --log-disconnections string          [OPTIONAL] Log session disconnections. (default "false")
      --log-duration string                [OPTIONAL] Log the duration of each completed statement. (default "false")
      --log-error-verbosity string         [OPTIONAL] Controls the amount of detail written in the server log for each message that is logged. Options: DEFAULT, TERSE, VERBOSE. (default "DEFAULT")
      --log-statement string               [OPTIONAL] Log all statements or specific types of statements. Options: NONE, DDL, MOD, ALL. (default "NONE")
      --log-min-error-statement string     [OPTIONAL] Minimum error severity for logging the statement that caused it. Options: ERROR. (default "ERROR")
      --log-line-prefix string             [OPTIONAL] A printf-style format string for log line prefixes. (default "%m :%r :%u @ %d :[%p] :")
  -h, --help                               help for enable
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
      --host string           YugabyteDB Aeon Api hostname
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster db-query-logging](ybm_cluster_db-query-logging.md)	 - Configure Database Query Logging for your Cluster.

