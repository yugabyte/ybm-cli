## ybm cluster node

Manage nodes for a cluster

### Synopsis

Manage nodes for a cluster

```
ybm cluster node [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for node
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

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm cluster node list](ybm_cluster_node_list.md)	 - List nodes for a cluster

