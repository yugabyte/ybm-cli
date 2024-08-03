## ybm cluster network allow-list

Manage network allow-list operations for a cluster

### Synopsis

Manage network allow-list operations for a cluster

```
ybm cluster network allow-list [flags]
```

### Options

```
  -h, --help   help for allow-list
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

* [ybm cluster network](ybm_cluster_network.md)	 - Manage network operations
* [ybm cluster network allow-list assign](ybm_cluster_network_allow-list_assign.md)	 - Assign resources(e.g. network allow lists) to clusters
* [ybm cluster network allow-list unassign](ybm_cluster_network_allow-list_unassign.md)	 - Unassign resources(e.g. network allow lists) to clusters

