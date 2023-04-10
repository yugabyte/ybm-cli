## ybm cluster network allow-list unassign

Unassign resources(e.g. network allow lists) to clusters

### Synopsis

Unassign resources(e.g. network allow lists) to clusters

```
ybm cluster network allow-list unassign [flags]
```

### Options

```
  -h, --help                        help for unassign
      --network-allow-list string   [REQUIRED] The name of the network allow list to be unassigned.
```

### Options inherited from parent commands

```
  -a, --apiKey string         YBM Api Key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster network allow-list](ybm_cluster_network_allow-list.md)	 - Manage network allow-list operations for a cluster

