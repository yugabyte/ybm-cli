## ybm dr create

Create DR for a cluster

### Synopsis

Create DR for a cluster

```
ybm dr create [flags]
```

### Options

```
      --databases stringArray        [REQUIRED] Databases to be replicated. Please provide a comma separated list of database names <db-name-1>,<db-name-2>.
      --dr-name string               [REQUIRED] Name of the DR configuration.
  -h, --help                         help for create
      --source-cluster-name string   [REQUIRED] Source cluster in the DR configuration.
      --target-cluster-name string   [REQUIRED] Target cluster in the DR configuration.
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

* [ybm dr](ybm_dr.md)	 - Manage DR for a cluster.

