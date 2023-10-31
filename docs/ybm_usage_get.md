## ybm usage get

View billing usage data available for the account in YugabyteDB Managed

### Synopsis

View billing usage data available for the account in YugabyteDB Managed

```
ybm usage get [flags]
```

### Options

```
      --cluster-name stringArray   [REQUIRED] Cluster names. Multiple names can be specified by using multiple --cluster-name arguments.
      --end string                 [REQUIRED] End date in RFC3339 format (e.g., '2023-09-30T23:59:59.999Z') or 'yyyy-MM-dd' format (e.g., '2023-09-30').
  -h, --help                       help for get
      --output-file string         [OPTIONAL] Output filename.
      --output-format string       [OPTIONAL] Output format. Possible values: csv, json. (default "csv")
      --start string               [REQUIRED] Start date in RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01').
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

* [ybm usage](ybm_usage.md)	 - Billing usage for the account in YugabyteDB Managed

