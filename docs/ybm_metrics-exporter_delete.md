## ybm metrics-exporter delete

Delete Metrics Exporter Config

### Synopsis

Delete Metrics Exporter Config

```
ybm metrics-exporter delete [flags]
```

### Options

```
      --config-name string   [REQUIRED] The name of the metrics exporter configuration
  -f, --force                Bypass the prompt for non-interactive usage
  -h, --help                 help for delete
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

* [ybm metrics-exporter](ybm_metrics-exporter.md)	 - Manage Metrics Exporter

