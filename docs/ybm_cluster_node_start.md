## ybm cluster node start

Start a cluster node

### Synopsis

Start a cluster node

```
ybm cluster node start [flags]
```

### Options

```
      --cluster-name string   [REQUIRED] The name of the cluster to get details.
  -h, --help                  help for start
      --node-name string      [REQUIRED] The name of the node to stop.
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

* [ybm cluster node](ybm_cluster_node.md)	 - Manage nodes for a cluster

