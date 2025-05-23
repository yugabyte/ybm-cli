## ybm metrics-exporter

Manage Metrics Exporter

### Synopsis

Manage Metrics Exporter

```
ybm metrics-exporter [flags]
```

### Options

```
  -h, --help   help for metrics-exporter
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!
* [ybm metrics-exporter assign](ybm_metrics-exporter_assign.md)	 - Associate Metrics Exporter Config with Cluster
* [ybm metrics-exporter pause](ybm_metrics-exporter_pause.md)	 - Stop Metrics Exporter
* [ybm metrics-exporter unassign](ybm_metrics-exporter_unassign.md)	 - Unassign Metrics Exporter Config from Cluster

