## ybm cluster network endpoint

Manage network endpoints for a cluster

### Synopsis

Manage network endpoints for a cluster

```
ybm cluster network endpoint [flags]
```

### Options

```
  -h, --help   help for endpoint
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
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

* [ybm cluster network](ybm_cluster_network.md)	 - Manage network operations
* [ybm cluster network endpoint create](ybm_cluster_network_endpoint_create.md)	 - Create a new network endpoint for a cluster
* [ybm cluster network endpoint delete](ybm_cluster_network_endpoint_delete.md)	 - Delete a network endpoint for a cluster
* [ybm cluster network endpoint describe](ybm_cluster_network_endpoint_describe.md)	 - Describe a network endpoint for a cluster
* [ybm cluster network endpoint list](ybm_cluster_network_endpoint_list.md)	 - List network endpoints for a cluster
* [ybm cluster network endpoint update](ybm_cluster_network_endpoint_update.md)	 - Update a network endpoint for a cluster

