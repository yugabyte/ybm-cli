## ybm metrics-exporter unassign

Unassign Metrics Exporter Config from Cluster

### Synopsis

Unassign Metrics Exporter Config from Cluster

```
ybm metrics-exporter unassign [flags]
```

### Options

```
      --cluster-name string   [REQUIRED] The name of the cluster
  -h, --help                  help for unassign
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

* [ybm metrics-exporter](ybm_metrics-exporter.md)	 - Manage Metrics Exporter

